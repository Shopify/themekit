package kit

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
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
	httpClient *http.Client
	filter     eventFilter
}

type apiResponse struct {
	code int
	body []byte
	err  error
}

// EventType is an enum of event types to compare agains event.Type()
type EventType int

func (e EventType) String() string {
	switch e {
	case Update:
		return "Update"
	case Remove:
		return "Remove"
	default:
		return "Unknown"
	}
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

const (
	// Update specifies that an AssetEvent is an update event.
	Update EventType = iota
	// Remove specifies that an AssetEvent is an delete event.
	Remove
)

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
func (t ThemeClient) AssetList() []theme.Asset {
	queryBuilder := func(path string) string {
		return path
	}

	resp := t.query(queryBuilder)
	if resp.err != nil {
		Errorf(resp.err.Error())
	}

	if resp.code >= 400 && resp.code < 500 {
		Errorf("Server responded with HTTP %d; please check your credentials.", resp.code)
		return []theme.Asset{}
	}
	if resp.code >= 500 {
		Errorf("Server responded with HTTP %d; try again in a few minutes.", resp.code)
		return []theme.Asset{}
	}

	var assets map[string][]theme.Asset
	err := json.Unmarshal(resp.body, &assets)
	if err != nil {
		Errorf(err.Error())
		return []theme.Asset{}
	}

	sort.Sort(theme.ByAsset(assets["assets"]))

	return t.filter.filterAssets(ignoreCompiledAssets(assets["assets"]))
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
	queryBuilder := func(path string) string {
		return fmt.Sprintf("%s&asset[key]=%s", path, filename)
	}

	resp := t.query(queryBuilder)
	if resp.err != nil {
		return theme.Asset{}, resp.err
	}
	if resp.code >= 400 {
		return theme.Asset{}, NonFatalNetworkError{Code: resp.code, Verb: "GET", Message: "not found"}
	}
	var asset map[string]theme.Asset
	err := json.Unmarshal(resp.body, &asset)
	if err != nil {
		return theme.Asset{}, err
	}

	return asset["asset"], nil
}

// CreateTheme will create a unpublished new theme on your shopify store and then
// return a new theme client with the configuration of the new client.
func CreateTheme(name, zipLocation string) (ThemeClient, error) {
	config, _ := NewConfiguration()
	client, err := NewThemeClient(config)
	if err != nil {
		return client, err
	}

	var themeEvent apiThemeEvent
	retries := 0
	data, err := json.Marshal(map[string]theme.Theme{
		"theme": {Name: name, Source: zipLocation, Role: "unpublished"},
	})
	if err != nil {
		return client, err
	}

	for retries < createThemeMaxRetries {
		themeEvent, err = client.sendData("POST", fmt.Sprintf("%s/themes.json", config.AdminURL()), data)
		if err != nil {
			return client, err
		} else if !themeEvent.Successful() {
			retries++
		} else {
			break
		}
		Logf(themeEvent.String())
	}

	if retries >= createThemeMaxRetries {
		return client, fmt.Errorf("'%s' cannot be retrieved from Github.", zipLocation)
	}

	for !client.isDoneProcessing(themeEvent.ThemeID) {
		time.Sleep(250 * time.Millisecond)
	}

	config.ThemeID = fmt.Sprintf("%d", themeEvent.ThemeID)

	return client, err
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
func (t ThemeClient) Perform(asset AssetEvent) {
	if t.filter.matchesFilter(asset.Asset().Key) {
		return
	}

	var event string
	switch asset.Type() {
	case Update:
		event = "PUT"
	case Remove:
		event = "DELETE"
	}

	resp, err := t.request(asset, event)
	Logf(newAPIAssetEvent(resp, asset, err).String())
}

func (t ThemeClient) query(queryBuilder func(path string) string) apiResponse {
	path := fmt.Sprintf("%s?fields=key,attachment,value", t.config.AssetPath())
	path = queryBuilder(path)

	req, err := http.NewRequest("GET", path, nil)
	if err != nil {
		return apiResponse{err: err}
	}

	t.config.AddHeaders(req)
	resp, err := t.httpClient.Do(req)
	if err != nil {
		return apiResponse{err: err}
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	return apiResponse{code: resp.StatusCode, body: body, err: err}
}

func (t ThemeClient) sendData(method, path string, body []byte) (result apiThemeEvent, err error) {
	req, err := http.NewRequest(method, path, bytes.NewBuffer(body))
	if err != nil {
		return apiThemeEvent{}, err
	}
	t.config.AddHeaders(req)
	resp, err := t.httpClient.Do(req)
	if result = newAPIThemeEvent(resp, err); result.Successful() {
		defer resp.Body.Close()
	}
	return result, nil
}

func (t ThemeClient) request(event AssetEvent, method string) (*http.Response, error) {
	path := t.config.AssetPath()
	data := map[string]theme.Asset{"asset": event.Asset()}

	encoded, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(method, path, bytes.NewBuffer(encoded))

	if err != nil {
		return nil, err
	}

	t.config.AddHeaders(req)
	return t.httpClient.Do(req)
}

func (t ThemeClient) isDoneProcessing(themeID int64) bool {
	path := fmt.Sprintf("%s/themes/%d.json", t.config.AdminURL(), themeID)
	themeEvent, err := t.sendData("GET", path, []byte{})
	return err == nil && themeEvent.Previewable
}

func newHTTPClient(config Configuration) (client *http.Client, err error) {
	client = &http.Client{
		Timeout: config.Timeout,
	}

	if len(config.Proxy) > 0 {
		Warnf("Proxy URL detected from Configuration:", config.Proxy)
		Warnf("SSL Certificate Validation will be disabled!")
		proxyURL, err := url.Parse(config.Proxy)
		if err != nil {
			return client, fmt.Errorf("Proxy configuration invalid:", err)
		}
		client.Transport = &http.Transport{Proxy: http.ProxyURL(proxyURL), TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}
	}
	return
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
