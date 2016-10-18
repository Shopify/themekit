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

	"github.com/Shopify/themekit/theme"
)

type httpClient struct {
	client *http.Client
	config Configuration
}

func newHTTPClient(config Configuration) (*httpClient, error) {
	client := &httpClient{
		client: &http.Client{Timeout: config.Timeout},
		config: config,
	}

	if len(config.Proxy) > 0 {
		Warnf("Proxy URL detected from Configuration:", config.Proxy)
		Warnf("SSL Certificate Validation will be disabled!")
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

func (client *httpClient) AssetQuery(event EventType, query map[string]string) (*shopifyResponse, error) {
	path := fmt.Sprintf("%s?fields=key,attachment,value", client.AssetPath())
	for key, value := range query {
		path += "&" + key + "=" + value
	}
	return client.sendRequest(event, path, nil)
}

func (client *httpClient) NewTheme(name, source string) (*shopifyResponse, error) {
	return client.sendJSON(Update, client.AssetPath(), map[string]interface{}{
		"theme": theme.Theme{Name: name, Source: source, Role: "unpublished"},
	})
}

func (client *httpClient) GetTheme(themeID int64) (*shopifyResponse, error) {
	return client.sendRequest(Retrieve, client.ThemePath(themeID), nil)
}

func (client *httpClient) AssetAction(event AssetEvent) (*shopifyResponse, error) {
	return client.sendJSON(event.Type(), client.AssetPath(), map[string]interface{}{
		"asset": event.Asset(),
	})
}

func (client *httpClient) newRequest(event EventType, urlStr string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequest(event.ToMethod(), urlStr, body)
	if err != nil {
		return nil, err
	}

	req.Header.Add("X-Shopify-Access-Token", client.config.Password)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")
	req.Header.Add("User-Agent", fmt.Sprintf("go/themekit (%s; %s)", runtime.GOOS, runtime.GOARCH))

	return req, nil
}

func (client *httpClient) sendJSON(event EventType, urlStr string, body map[string]interface{}) (*shopifyResponse, error) {
	data, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	return client.sendRequest(event, client.AssetPath(), bytes.NewBuffer(data))
}

func (client *httpClient) sendRequest(event EventType, urlStr string, body io.Reader) (*shopifyResponse, error) {
	req, err := client.newRequest(event, urlStr, body)
	if err != nil {
		return nil, err
	}
	resp, err := client.client.Do(req)
	return newShopifyResponse(event, resp, err)
}
