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
	"strings"
	"time"
)

var apiLimit = newRateLimiter(time.Second / 2)

type httpClient struct {
	client *http.Client
	config *Configuration
}

type requestType int

const (
	themeRequest requestType = iota
	assetRequest
	listRequest
)

func newHTTPClient(config *Configuration) (*httpClient, error) {
	client := &httpClient{
		client: &http.Client{Timeout: config.Timeout},
		config: config,
	}

	if len(config.Proxy) > 0 {
		Print(YellowText(fmt.Sprintf("Proxy URL detected from Configuration: %s", config.Proxy)))
		Print(YellowText("SSL Certificate Validation will be disabled!"))
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
	rtype := listRequest
	if len(query) > 0 {
		rtype = assetRequest
	}
	return client.sendRequest(rtype, event, path, nil)
}

func (client *httpClient) NewTheme(name, source string) (*ShopifyResponse, Error) {
	return client.sendJSON(themeRequest, Create, client.ThemesPath(), map[string]interface{}{
		"theme": Theme{Name: name, Source: source, Role: "unpublished"},
	})
}

func (client *httpClient) GetTheme(themeID int64) (*ShopifyResponse, Error) {
	return client.sendRequest(themeRequest, Retrieve, client.ThemePath(themeID), nil)
}

func (client *httpClient) AssetAction(event EventType, asset Asset) (*ShopifyResponse, Error) {
	resp, _ := client.sendJSON(assetRequest, event, client.AssetPath(), map[string]interface{}{
		"asset": asset,
	})
	// If there were any errors the asset is nil so lets set it and reformat errors
	resp.Asset = asset
	return resp, resp.Error()
}

func (client *httpClient) newRequest(event EventType, urlStr string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequest(event.toMethod(), urlStr, body)
	if err != nil {
		return nil, err
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
		return newShopifyResponse(rtype, event, urlStr, nil, err)
	}
	return client.sendRequest(rtype, event, urlStr, bytes.NewBuffer(data))
}

func (client *httpClient) sendRequest(rtype requestType, event EventType, urlStr string, body io.Reader) (*ShopifyResponse, Error) {
	if client.config.ReadOnly && event != Retrieve {
		return newShopifyResponse(rtype, event, urlStr, nil, fmt.Errorf("Theme is read only"))
	}

	req, err := client.newRequest(event, urlStr, body)
	if err != nil {
		return newShopifyResponse(rtype, event, urlStr, nil, err)
	}

	apiLimit.Wait()

	resp, respErr := client.client.Do(req)
	return newShopifyResponse(rtype, event, urlStr, resp, respErr)
}
