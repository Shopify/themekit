package kit

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/Shopify/themekit/theme"
)

const createThemeMaxRetries int = 3

// ThemeClient is the interactor with the shopify server. All actions are processed
// with the client.
type ThemeClient struct {
	config     Configuration
	httpClient *httpClient
	filter     eventFilter
}

type apiResponse struct {
	code int
	body []byte
	err  error
}

// NonFatalNetworkError is an error for describing an asset request that failed
// but is not critical. For instance updating a file that does not exist.
type NonFatalNetworkError struct {
	Code    int
	Verb    string
	Message string
}

func (e NonFatalNetworkError) Error() string {
	return fmt.Sprintf("%d %s %s", e.Code, e.Verb, e.Message)
}

// NewThemeClient will build a new theme client from a configuration and a theme event
// channel. The channel is used for logging all events. The configuration specifies how
// the client will behave.
func NewThemeClient(config Configuration) (ThemeClient, error) {
	httpClient, err := newHTTPClient(config)
	if err != nil {
		return ThemeClient{}, err
	}

	filter, err := newEventFilter(config.Directory, config.IgnoredFiles, config.Ignores)
	if err != nil {
		return ThemeClient{}, err
	}

	return ThemeClient{
		config:     config,
		httpClient: httpClient,
		filter:     filter,
	}, nil
}

// GetConfiguration will return the clients built config. This is useful for grabbing
// things like urls and domains.
func (t ThemeClient) GetConfiguration() Configuration {
	return t.config
}

// NewFileWatcher creates a new filewatcher using the theme clients file filter
func (t ThemeClient) NewFileWatcher(notifyFile string) (chan AssetEvent, error) {
	newForeman := newForeman(newLeakyBucket(t.config.BucketSize, t.config.RefillRate, 1))
	if len(notifyFile) > 0 {
		newForeman.OnIdle = func() {
			os.Create(notifyFile)
			os.Chtimes(notifyFile, time.Now(), time.Now())
		}
	}
	var err error
	newForeman.JobQueue, err = newFileWatcher(t.config.Directory, true, t.filter)
	if err != nil {
		return newForeman.JobQueue, err
	}
	newForeman.Restart()
	return newForeman.WorkerQueue, nil
}

// AssetList will return a slice of remote assets from the shopify servers. The
// assets are sorted and any ignored files based on your config are filtered out.
func (t ThemeClient) AssetList() ([]theme.Asset, error) {
	resp, err := t.httpClient.AssetQuery(Retrieve, map[string]string{})
	if err != nil {
		return []theme.Asset{}, err
	} else if !resp.Successful() {
		return []theme.Asset{}, fmt.Errorf(resp.String())
	}
	sort.Sort(theme.ByAsset(resp.Assets))
	return t.filter.filterAssets(ignoreCompiledAssets(resp.Assets)), nil
}

// LocalAssets will return a slice of assets from the local disk. The
// assets are filtered based on your config.
func (t ThemeClient) LocalAssets() ([]theme.Asset, error) {
	dir := fmt.Sprintf("%s%s", t.config.Directory, string(filepath.Separator))
	assets, err := theme.LoadAssetsFromDirectory(dir, t.filter.matchesFilter)
	if err != nil {
		return assets, err
	}
	return assets, nil
}

// LoadAsset will load a single local asset on disk. It will return an error if there
// is a problem loading the asset.
func (t ThemeClient) LocalAsset(filename string) (theme.Asset, error) {
	return theme.LoadAsset(t.config.Directory, filename)
}

// Asset will load up a single remote asset from the remote shopify servers.
func (t ThemeClient) Asset(filename string) (theme.Asset, error) {
	resp, err := t.httpClient.AssetQuery(Retrieve, map[string]string{"asset[key]": filename})
	return resp.Asset, err
}

// CreateTheme will create a unpublished new theme on your shopify store and then
// return a new theme client with the configuration of the new client.
func CreateTheme(name, zipLocation string) (ThemeClient, error) {
	config, _ := NewConfiguration()
	client, err := NewThemeClient(config)
	if err != nil {
		return client, err
	}

	var resp *shopifyResponse
	retries := 0
	for retries < createThemeMaxRetries {
		resp, err = client.httpClient.NewTheme(name, zipLocation)
		if err != nil {
			return client, err
		} else if !resp.Successful() {
			retries++
			if retries >= createThemeMaxRetries {
				return client, fmt.Errorf("'%s' cannot be retrieved from Github.", zipLocation)
			}
		} else {
			break
		}
		Logf(resp.String())
	}

	for !client.isDoneProcessing(resp.Theme.ID) {
		time.Sleep(250 * time.Millisecond)
	}

	config.ThemeID = fmt.Sprintf("%d", resp.Theme.ID)

	return client, err
}

func (t ThemeClient) isDoneProcessing(themeID int64) bool {
	resp, err := t.httpClient.GetTheme(themeID)
	return err == nil && resp.Theme.Previewable
}

// Process will create a new throttler and return the jobqueue. You can then send
// asset events to the channel and they will be performed. If you close the job
// queue, then the worker queue will be closed when it is finished, then the done
// channel will be closed. This is a good way of knowing when your jobs are done
// processing.
func (t ThemeClient) Process(wg *sync.WaitGroup) chan AssetEvent {
	newForeman := newForeman(newLeakyBucket(t.config.BucketSize, t.config.RefillRate, 1))
	var processWaitGroup sync.WaitGroup
	go func() {
		for {
			job, more := <-newForeman.WorkerQueue
			if more {
				processWaitGroup.Add(1)
				go func() {
					t.Perform(job)
					processWaitGroup.Done()
				}()
			} else {
				processWaitGroup.Wait()
				wg.Done()
				return
			}
		}
	}()
	return newForeman.JobQueue
}

// Perform will send an http request to the shopify servers based on the asset event.
// if it is an update it will post to the server and if it is a remove it will DELETE
// to the server. Any errors will be outputted to the event log.
func (t ThemeClient) Perform(asset AssetEvent) error {
	if t.filter.matchesFilter(asset.Asset().Key) {
		return fmt.Errorf(YellowText(fmt.Sprintf("Asset %s filtered based on ignore patterns", asset.Asset().Key)))
	}
	resp, err := t.httpClient.AssetAction(asset)
	if err != nil {
		return err
	}
	Logf(resp.String())
	return nil
}

func ignoreCompiledAssets(assets []theme.Asset) []theme.Asset {
	newSize := 0
	results := make([]theme.Asset, len(assets))
	isCompiled := func(a theme.Asset, rest []theme.Asset) bool {
		for _, other := range rest {
			if strings.Contains(other.Key, a.Key) {
				return true
			}
		}
		return false
	}
	for index, asset := range assets {
		if !isCompiled(asset, assets[index+1:]) {
			results[newSize] = asset
			newSize++
		}
	}
	return results[:newSize]
}
