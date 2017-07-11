package kit

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

type httpClient struct {
	client *http.Client
	config *Configuration
	limit  *rateLimiter
}

type requestType int

const (
	themeRequest requestType = iota
	assetRequest
	listRequest

	assetDataFields = "key,updated_at"
)

func newHTTPClient(config *Configuration) (*httpClient, error) {
	client := &httpClient{
		client: &http.Client{Timeout: config.Timeout},
		config: config,
		limit:  rateLimitFor(config.Domain),
	}

	if len(config.Proxy) > 0 {
		proxyURL, err := url.Parse(config.Proxy)
		if err != nil {
			return client, err
		}
		client.client.Transport = &http.Transport{
			Proxy:           http.ProxyURL(proxyURL),
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
	}

	return client, nil
}

// AdminURL will return the url to the shopify admin.
func (client *httpClient) AdminURL() string {
	adminURL := fmt.Sprintf("%s/admin", client.config.Domain)
	if !client.config.IsLive() {
		if themeID, err := strconv.ParseInt(client.config.ThemeID, 10, 64); err == nil {
			adminURL = fmt.Sprintf("%s/themes/%d", adminURL, themeID)
		}
	}
	parsedURL, _ := url.Parse(adminURL)
	parsedURL.Scheme = "https"
	// for testing, because otherwise the domain should not be localhost or http
	if strings.HasPrefix(client.config.Domain, "http://127.0.0.1:") {
		parsedURL.Scheme = "http"
	}
	return parsedURL.String()
}

// AssetPath will return the assets endpoint in the admin section of shopify.
func (client *httpClient) AssetPath(query map[string]string) string {
	if len(query) == 0 {
		return fmt.Sprintf("%s/assets.json", client.AdminURL())
	}

	queryParams := url.Values{}
	for key, value := range query {
		queryParams.Set(key, value)
	}
	return fmt.Sprintf("%s/assets.json?%s", client.AdminURL(), queryParams.Encode())
}

// ThemesPath will return the endpoint of the themes interactions.
func (client *httpClient) ThemesPath() string {
	return fmt.Sprintf("%s/themes.json", client.AdminURL())
}

// ThemePath will return the endpoint of the a single theme.
func (client *httpClient) ThemePath(themeID int64) string {
	return fmt.Sprintf("%s/themes/%d.json", client.AdminURL(), themeID)
}

// NewTheme will create a request to create a new theme with the proviced name and
// create it with the source provided
func (client *httpClient) NewTheme(name, source string) (*ShopifyResponse, Error) {
	req, err := newShopifyRequest(client.config, themeRequest, Create, client.ThemesPath())
	if err != nil {
		return newShopifyResponse(req, nil, err)
	}

	err = req.setJSONBody(map[string]interface{}{"theme": Theme{Name: name, Source: source, Role: "unpublished"}})
	if err != nil {
		return newShopifyResponse(req, nil, err)
	}

	return client.sendRequest(req)
}

// GetTheme will load the theme data for the provided theme id
func (client *httpClient) GetTheme(themeID int64) (*ShopifyResponse, Error) {
	req, err := newShopifyRequest(client.config, themeRequest, Retrieve, client.ThemePath(themeID))
	if err != nil {
		return newShopifyResponse(req, nil, err)
	}
	return client.sendRequest(req)
}

// AssetList will return a shopify respinse with []Assets defined however none of
// the assets will have a body. Those will have to be requested separately
func (client *httpClient) AssetList() (*ShopifyResponse, Error) {
	req, err := newShopifyRequest(client.config, listRequest, Retrieve, client.AssetPath(
		map[string]string{"fields": assetDataFields},
	))
	if err != nil {
		return newShopifyResponse(req, nil, err)
	}
	return client.sendRequest(req)
}

func (client *httpClient) GetAssetInfo(filename string) (*ShopifyResponse, Error) {
	path := client.AssetPath(map[string]string{
		"asset[key]": filename,
		"fields":     assetDataFields,
	})
	req, err := newShopifyRequest(client.config, assetRequest, Retrieve, path)
	if err != nil {
		return newShopifyResponse(req, nil, err)
	}

	return client.sendRequest(req)
}

func (client *httpClient) GetAsset(filename string) (*ShopifyResponse, Error) {
	path := client.AssetPath(map[string]string{"asset[key]": filename})
	req, err := newShopifyRequest(client.config, assetRequest, Retrieve, path)
	if err != nil {
		return newShopifyResponse(req, nil, err)
	}

	return client.sendRequest(req)
}

func (client *httpClient) AssetAction(event EventType, asset Asset) (*ShopifyResponse, Error) {
	req, err := newShopifyRequest(client.config, assetRequest, event, client.AssetPath(nil))
	if err != nil {
		return newShopifyResponse(req, nil, err)
	}

	err = req.setJSONBody(map[string]interface{}{"asset": asset})
	if err != nil {
		return newShopifyResponse(req, nil, err)
	}

	resp, err := client.sendRequest(req)
	// If there were any errors the asset is nil so lets set it and reformat errors
	if err != nil || resp.Asset.Key == "" {
		resp.Asset = asset
	}
	return resp, resp.Error()
}

func (client *httpClient) AssetActionStrict(event EventType, asset Asset, version string) (*ShopifyResponse, Error) {
	req, err := newShopifyRequest(client.config, assetRequest, event, client.AssetPath(nil))
	if err != nil {
		return newShopifyResponse(req, nil, err)
	}

	err = req.setJSONBody(map[string]interface{}{"asset": asset})
	if err != nil {
		return newShopifyResponse(req, nil, err)
	}

	req.Header.Add("If-Unmodified-Since", version)

	resp, err := client.sendRequest(req)
	// If there were any errors the asset is nil so lets set it and reformat errors
	if err != nil || resp.Asset.Key == "" {
		resp.Asset = asset
	}
	return resp, resp.Error()
}

func (client *httpClient) sendRequest(req *shopifyRequest) (*ShopifyResponse, Error) {
	if client.config.ReadOnly && req.event != Retrieve {
		return newShopifyResponse(req, nil, fmt.Errorf("Theme is read only"))
	}
	client.limit.Wait()
	resp, respErr := client.client.Do(req.Request)
	return newShopifyResponse(req, resp, respErr)
}
