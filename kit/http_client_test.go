package kit

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Shopify/themekit/kittest"
)

func TestNewHttpClient(t *testing.T) {
	config, _ := NewConfiguration()
	config.Proxy = "://abc!21@"
	client, err := newHTTPClient(config)
	assert.NotNil(t, err)

	config, _ = NewConfiguration()
	config.Proxy = "http://localhost:3000"
	client, err = newHTTPClient(config)
	assert.Nil(t, err)
	assert.NotNil(t, client.client.Transport)
}

func TestAdminURL(t *testing.T) {
	client, _ := newHTTPClient(&Configuration{Domain: "domain", ThemeID: "123"})
	assert.Equal(t, "https://domain/admin/themes/123", client.AdminURL())
	client.config.ThemeID = "live"
	assert.Equal(t, "https://domain/admin", client.AdminURL())
	client, _ = newHTTPClient(&Configuration{Domain: "domain", ThemeID: "themeid"})
	assert.Equal(t, "https://domain/admin", client.AdminURL())
}

func TestAssetPath(t *testing.T) {
	client, _ := newHTTPClient(&Configuration{Domain: "domain", ThemeID: "live"})
	assert.Equal(t, "https://domain/admin/assets.json", client.AssetPath(nil))
}

func TestThemesPath(t *testing.T) {
	client, _ := newHTTPClient(&Configuration{Domain: "domain", ThemeID: "live"})
	assert.Equal(t, "https://domain/admin/themes.json", client.ThemesPath())
}

func TestThemePath(t *testing.T) {
	client, _ := newHTTPClient(&Configuration{Domain: "domain", ThemeID: "live"})
	assert.Equal(t, "https://domain/admin/themes/456.json", client.ThemePath(456))
}

func TestAssetList(t *testing.T) {
	server := kittest.NewTestServer()
	defer server.Close()
	client, _ := newHTTPClient(&Configuration{Domain: server.URL, ThemeID: "123"})
	resp, err := client.AssetList()
	assert.Nil(t, err)
	assert.Equal(t, listRequest, resp.Type)
	assert.Equal(t, 2, len(resp.Assets))
}

func TestGetAsset(t *testing.T) {
	server := kittest.NewTestServer()
	defer server.Close()
	client, _ := newHTTPClient(&Configuration{Domain: server.URL, ThemeID: "123"})

	resp, err := client.GetAsset("file.txt")
	assert.Nil(t, err)
	assert.Equal(t, assetRequest, resp.Type)
	assert.Equal(t, "assets/hello.txt", resp.Asset.Key)
}

func TestNewTheme(t *testing.T) {
	server := kittest.NewTestServer()
	defer server.Close()
	client, _ := newHTTPClient(&Configuration{Domain: server.URL, ThemeID: "123"})
	resp, err := client.NewTheme("name", "source")
	assert.Nil(t, err)
	assert.Equal(t, themeRequest, resp.Type)
	assert.Equal(t, "timberland", resp.Theme.Name)
	assert.Equal(t, 1, len(server.Requests))
	assert.Equal(t, "POST", server.Requests[0].Method)
}

func TestGetTheme(t *testing.T) {
	server := kittest.NewTestServer()
	defer server.Close()
	client, _ := newHTTPClient(&Configuration{Domain: server.URL, ThemeID: "123"})
	resp, err := client.GetTheme(123)
	assert.Nil(t, err)
	assert.Equal(t, themeRequest, resp.Type)
	assert.Equal(t, "timberland", resp.Theme.Name)
	assert.Equal(t, 1, len(server.Requests))
	assert.Equal(t, "GET", server.Requests[0].Method)
}

func TestAssetAction(t *testing.T) {
	server := kittest.NewTestServer()
	defer server.Close()
	client, _ := newHTTPClient(&Configuration{Domain: server.URL, ThemeID: "123"})
	resp, err := client.AssetAction(Update, Asset{Key: "key", Value: "value"})
	assert.Nil(t, err)
	assert.Equal(t, assetRequest, resp.Type)
	assert.Equal(t, "assets/hello.txt", resp.Asset.Key)
	assert.Equal(t, 1, len(server.Requests))
	assert.Equal(t, "PUT", server.Requests[0].Method)
}

func TestAssetActionStrict(t *testing.T) {
	server := kittest.NewTestServer()
	defer server.Close()
	client, _ := newHTTPClient(&Configuration{Domain: server.URL, ThemeID: "123"})
	resp, err := client.AssetActionStrict(Update, Asset{Key: "key", Value: "value"}, "version")
	assert.Nil(t, err)
	assert.Equal(t, assetRequest, resp.Type)
	assert.Equal(t, "assets/hello.txt", resp.Asset.Key)
	assert.Equal(t, 1, len(server.Requests))
	assert.Equal(t, "PUT", server.Requests[0].Method)
	assert.Equal(t, server.Requests[0].Header.Get("If-Unmodified-Since"), "version")

	_, err = client.AssetActionStrict(Update, Asset{Key: "nope", Value: "value"}, "version")
	assert.NotNil(t, err)
}

func TestSendRequest(t *testing.T) {
	server := kittest.NewTestServer()
	defer server.Close()
	client, _ := newHTTPClient(&Configuration{Domain: server.URL, ThemeID: "123"})
	req := newShopifyRequest(client.config, assetRequest, Update, client.AssetPath(map[string]string{"asset[key]": "test"}))
	resp, err := client.sendRequest(req)
	assert.Nil(t, err)
	assert.Equal(t, assetRequest, resp.Type)
	assert.Equal(t, "PUT", server.Requests[0].Method)

	client.config.ReadOnly = true
	_, err = client.sendRequest(req)
	assert.NotNil(t, err)
	assert.True(t, strings.Contains(err.Error(), "Theme is read only"))
}
