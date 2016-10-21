package kit

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const createThemeMaxRetries int = 3

// ThemeClient is the interactor with the shopify server. All actions are processed
// with the client.
type ThemeClient struct {
	config     Configuration
	httpClient *httpClient
	filter     eventFilter
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

	newClient := ThemeClient{
		config:     config,
		httpClient: httpClient,
		filter:     filter,
	}

	return newClient, nil
}

// GetConfiguration will return the clients built config. This is useful for grabbing
// things like urls and domains.
func (t ThemeClient) GetConfiguration() Configuration {
	return t.config
}

// NewFileWatcher creates a new filewatcher using the theme clients file filter
func (t ThemeClient) NewFileWatcher(notifyFile string, callback FileEventCallback) (*FileWatcher, error) {
	return newFileWatcher(t, t.config.Directory, true, t.filter, callback)
}

// AssetList will return a slice of remote assets from the shopify servers. The
// assets are sorted and any ignored files based on your config are filtered out.
func (t ThemeClient) AssetList() ([]Asset, Error) {
	resp, err := t.httpClient.AssetQuery(Retrieve, map[string]string{})
	if err != nil && err.Fatal() {
		return []Asset{}, err
	}
	sort.Sort(ByAsset(resp.Assets))
	return t.filter.filterAssets(ignoreCompiledAssets(resp.Assets)), err
}

// LocalAssets will return a slice of assets from the local disk. The
// assets are filtered based on your config.
func (t ThemeClient) LocalAssets() ([]Asset, error) {
	dir := fmt.Sprintf("%s%s", t.config.Directory, string(filepath.Separator))
	assets, err := LoadAssetsFromDirectory(dir, t.filter.matchesFilter)
	if err != nil {
		return assets, err
	}
	return assets, nil
}

// LocalAsset will load a single local asset on disk. It will return an error if there
// is a problem loading the asset.
func (t ThemeClient) LocalAsset(filename string) (Asset, error) {
	return LoadAsset(t.config.Directory, filename)
}

// Asset will load up a single remote asset from the remote shopify servers.
func (t ThemeClient) Asset(filename string) (Asset, Error) {
	resp, err := t.httpClient.AssetQuery(Retrieve, map[string]string{"asset[key]": filename})
	if err != nil {
		return Asset{}, err
	}
	return resp.Asset, nil
}

// CreateTheme will create a unpublished new theme on your shopify store and then
// return a new theme client with the configuration of the new client.
func CreateTheme(name, zipLocation string) (ThemeClient, error) {
	config, _ := NewConfiguration()
	err := config.validateNoThemeID()
	if err != nil {
		return ThemeClient{}, fmt.Errorf("Invalid options: %v", err)
	}

	client, err := NewThemeClient(config)
	if err != nil {
		return client, err
	}

	var resp *ShopifyResponse
	var respErr Error
	retries := 0
	for retries < createThemeMaxRetries {
		resp, respErr = client.httpClient.NewTheme(name, zipLocation)
		if respErr != nil {
			if respErr.Fatal() {
				return client, respErr
			}
		}

		retries++

		if resp.Successful() {
			Printf(
				"[%s]Successfully created theme '%s' with id of %s on shop %s",
				GreenText(resp.Code),
				BlueText(resp.Theme.Name),
				BlueText(resp.Theme.ID),
				YellowText(resp.Host),
			)
			break
		}

		if retries >= createThemeMaxRetries {
			return client, kitError{fmt.Errorf("Cannot create a theme. Please check log for errors.")}
		}
	}

	for !client.isDoneProcessing(resp.Theme.ID) {
		time.Sleep(250 * time.Millisecond)
	}

	client.config.ThemeID = fmt.Sprintf("%d", resp.Theme.ID)

	return client, nil
}

func (t ThemeClient) isDoneProcessing(themeID int64) bool {
	resp, err := t.httpClient.GetTheme(themeID)
	return err == nil && resp.Theme.Previewable
}

// CreateAsset will take an asset and will return  when the asset has been created.
// If there was an error, in the request then error will be defined otherwise the
//response will have the appropropriate data for usage.
func (t ThemeClient) CreateAsset(asset Asset) (*ShopifyResponse, Error) {
	return t.UpdateAsset(asset)
}

// UpdateAsset will take an asset and will return  when the asset has been updated.
// If there was an error, in the request then error will be defined otherwise the
//response will have the appropropriate data for usage.
func (t ThemeClient) UpdateAsset(asset Asset) (*ShopifyResponse, Error) {
	return t.Perform(asset, Update)
}

// DeleteAsset will take an asset and will return  when the asset has been deleted.
// If there was an error, in the request then error will be defined otherwise the
//response will have the appropropriate data for usage.
func (t ThemeClient) DeleteAsset(asset Asset) (*ShopifyResponse, Error) {
	return t.Perform(asset, Remove)
}

// Perform will take in any asset and event type, and return after the request has taken
// place
// If there was an error, in the request then error will be defined otherwise the
//response will have the appropropriate data for usage.
func (t ThemeClient) Perform(asset Asset, event EventType) (*ShopifyResponse, Error) {
	if t.filter.matchesFilter(asset.Key) {
		return &ShopifyResponse{}, kitError{fmt.Errorf(YellowText(fmt.Sprintf("Asset %s filtered based on ignore patterns", asset.Key)))}
	}
	return t.httpClient.AssetAction(event, asset)
}

func ignoreCompiledAssets(assets []Asset) []Asset {
	newSize := 0
	results := make([]Asset, len(assets))
	isCompiled := func(a Asset, rest []Asset) bool {
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
