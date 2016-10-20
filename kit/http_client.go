package kit

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"runtime"
	"strconv"
	"time"

	"github.com/Shopify/themekit/theme"
)

var apiLimit = newRateLimiter(time.Second / 2)

type httpClient struct {
	client *http.Client
	config Configuration
}

type requestType int

const (
	themeRequest requestType = iota
	assetRequest
	listRequest
)

func newHTTPClient(config Configuration) (*httpClient, error) {
	client := &httpClient{
		client: &http.Client{Timeout: config.Timeout},
		config: config,
	}

	if len(config.Proxy) > 0 {
		LogWarnf("Proxy URL detected from Configuration: %s", config.Proxy)
		LogWarn("SSL Certificate Validation will be disabled!")
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
	url := fmt.Sprintf("https://%s/admin", client.config.Domain)
	if !client.config.IsLive() {
		if themeID, err := strconv.ParseInt(client.config.ThemeID, 10, 64); err == nil {
			url = fmt.Sprintf("%s/themes/%d", url, themeID)
		}
	}
	return url
}

// AssetPath will return the assets endpoint in the admin section of shopify.
func (client *httpClient) AssetPath() string {
	return fmt.Sprintf("%s/assets.json", client.AdminURL())
}

// ThemesPath will return the endpoint of the themes interactions.
func (client *httpClient) ThemesPath() string {
	return fmt.Sprintf("%s/themes.json", client.AdminURL())
}

// ThemePath will return the endpoint of the a single theme.
func (client *httpClient) ThemePath(themeID int64) string {
	return fmt.Sprintf("%s/themes/%d.json", client.AdminURL(), themeID)
}

func (client *httpClient) AssetQuery(event EventType, query map[string]string) (*ShopifyResponse, Error) {
	path := fmt.Sprintf("%s?fields=key,attachment,value", client.AssetPath())
	for key, value := range query {
		path += "&" + key + "=" + value
	}
	rtype := assetRequest
	if len(query) > 0 {
		rtype = listRequest
	}
	return client.sendRequest(rtype, event, path, nil)
}

func (client *httpClient) NewTheme(name, source string) (*ShopifyResponse, Error) {
	return client.sendJSON(themeRequest, Update, client.ThemesPath(), map[string]interface{}{
		"theme": theme.Theme{Name: name, Source: source, Role: "unpublished"},
	})
}

func (client *httpClient) GetTheme(themeID int64) (*ShopifyResponse, Error) {
	return client.sendRequest(themeRequest, Retrieve, client.ThemePath(themeID), nil)
}

func (client *httpClient) AssetAction(event EventType, asset theme.Asset) (*ShopifyResponse, Error) {
	resp, _ := client.sendJSON(assetRequest, event, client.AssetPath(), map[string]interface{}{
		"asset": asset,
	})
	// If there were any errors the asset is nil so lets set it and reformat errors
	resp.Asset = asset
	return resp, resp.Error()
}

func (client *httpClient) newRequest(event EventType, urlStr string, body io.Reader) (*http.Request, Error) {
	req, err := http.NewRequest(event.toMethod(), urlStr, body)
	if err != nil {
		return nil, kitError{err}
	}

	req.Header.Add("X-Shopify-Access-Token", client.config.Password)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")
	req.Header.Add("User-Agent", fmt.Sprintf("go/themekit (%s; %s)", runtime.GOOS, runtime.GOARCH))

	return req, nil
}

func (client *httpClient) sendJSON(rtype requestType, event EventType, urlStr string, body map[string]interface{}) (*ShopifyResponse, Error) {
	data, err := json.Marshal(body)
	if err != nil {
		return nil, kitError{err}
	}
	return client.sendRequest(rtype, event, urlStr, bytes.NewBuffer(data))
}

func (client *httpClient) sendRequest(rtype requestType, event EventType, urlStr string, body io.Reader) (*ShopifyResponse, Error) {
	req, err := client.newRequest(event, urlStr, body)
	if err != nil {
		return nil, kitError{err}
	}
	apiLimit.Wait()
	resp, respErr := client.client.Do(req)
	return newShopifyResponse(rtype, event, resp, respErr)
}
