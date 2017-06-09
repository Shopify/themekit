package kit

import (
	"fmt"
	"strings"
	"time"
)

const createThemeMaxRetries int = 3

// ThemeClient is the interactor with the shopify server. All actions are processed
// with the client.
type ThemeClient struct {
	Config     *Configuration
	httpClient *httpClient
	filter     fileFilter
}

// NewThemeClient will build a new theme client from a configuration and a theme event
// channel. The channel is used for logging all events. The configuration specifies how
// the client will behave.
func NewThemeClient(config *Configuration) (ThemeClient, error) {
	httpClient, err := newHTTPClient(config)
	if err != nil {
		return ThemeClient{}, err
	}

	filter, err := newFileFilter(config.Directory, config.IgnoredFiles, config.Ignores)
	if err != nil {
		return ThemeClient{}, err
	}

	newClient := ThemeClient{
		Config:     config,
		httpClient: httpClient,
		filter:     filter,
	}

	return newClient, nil
}

// NewFileWatcher creates a new filewatcher using the theme clients file filter
func (t ThemeClient) NewFileWatcher(notifyFile string, callback FileEventCallback) (*FileWatcher, error) {
	return newFileWatcher(t, notifyFile, true, t.filter, callback)
}

// AssetList will return a slice of remote assets from the shopify servers. The
// assets are sorted and any ignored files based on your config are filtered out.
// The assets returned will not have any data, only ID and filenames. This is because
// fetching all the assets at one time is not a good idea.
func (t ThemeClient) AssetList() ([]Asset, Error) {
	resp, err := t.httpClient.AssetQuery(Retrieve, map[string]string{})
	if err != nil && err.Fatal() {
		return []Asset{}, err
	}
	return t.filter.filterAssets(resp.Assets), nil
}

// Asset will load up a single remote asset from the remote shopify servers.
func (t ThemeClient) Asset(filename string) (Asset, Error) {
	resp, err := t.httpClient.AssetQuery(Retrieve, map[string]string{"asset[key]": filename})
	if err != nil {
		return Asset{}, err
	}
	return resp.Asset, nil
}

// LocalAssets will return a slice of assets from the local disk. The
// assets are filtered based on your config. If not paths are passed to
// the function then all the local assets are returned. If you pass file names
// those assets will be loaded. If any of the file paths are directories, all
// of the directory's recursive assets will be returned.
func (t ThemeClient) LocalAssets(paths ...string) (assets []Asset, err error) {
	if paths == nil || len(paths) == 0 {
		assets, err = loadAssetsFromDirectory(t.Config.Directory, "", t.filter.matchesFilter)
		if err != nil {
			return assets, err
		}
	} else {
		assets = []Asset{}
		for _, path := range paths {
			asset, err := t.LocalAsset(path)
			if err == ErrAssetIsDir {
				dirAssets, err := loadAssetsFromDirectory(t.Config.Directory, path, t.filter.matchesFilter)
				if err != nil {
					return assets, err
				}
				assets = append(assets, dirAssets...)
			} else if err != nil {
				return assets, err
			} else {
				assets = append(assets, asset)
			}
		}
	}
	return t.filter.filterAssets(assets), nil
}

// LocalAsset will load a single local asset on disk. It will return an error if there
// is a problem loading the asset.
func (t ThemeClient) LocalAsset(filename string) (Asset, error) {
	return loadAsset(t.Config.Directory, filename)
}

// CreateTheme will create a unpublished new theme on your shopify store and then
// return a new theme client with the configuration of the new client.
func CreateTheme(name, zipLocation string) (ThemeClient, Theme, error) {
	config, _ := NewConfiguration()
	err := config.validateNoThemeID()
	if err != nil {
		return ThemeClient{}, Theme{}, fmt.Errorf("Invalid options: %v", err)
	}

	client, err := NewThemeClient(config)
	if err != nil {
		return client, Theme{}, err
	}

	var resp *ShopifyResponse
	var respErr error
	retries := 0
	for retries < createThemeMaxRetries {
		resp, respErr = client.httpClient.NewTheme(name, zipLocation)
		if resp != nil && resp.Successful() {
			break
		}

		retries++
		if retries >= createThemeMaxRetries {
			return client, Theme{}, kitError{fmt.Errorf("Cannot create a theme. Last error was: %s", respErr)}
		}
	}

	theme := resp.Theme
	for !theme.Previewable {
		resp, _ := client.httpClient.GetTheme(resp.Theme.ID)
		theme = resp.Theme
		time.Sleep(250 * time.Millisecond)
	}

	client.Config.ThemeID = fmt.Sprintf("%d", resp.Theme.ID)

	return client, theme, err
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
	return t.afterHooks(t.httpClient.AssetAction(event, asset))
}

func (t ThemeClient) afterHooks(resp *ShopifyResponse, err Error) (*ShopifyResponse, Error) {
	if resp.Code == 422 && strings.Contains(err.Error(), "Cannot overwrite generated asset") {
		// No need to check the error because if it fails then remove will be tried again.
		t.Perform(Asset{Key: resp.Asset.Key + ".liquid"}, Remove)
		resp, err = t.Perform(resp.Asset, Update)
	}

	return resp, err
}
