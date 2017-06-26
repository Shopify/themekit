package kit

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
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
	client := newTestHTTPClient()
	assert.Equal(t,
		fmt.Sprintf("https://%s/admin/themes/%v", client.config.Domain, client.config.ThemeID),
		client.AdminURL())
	client.config.ThemeID = "live"
	assert.Equal(t, fmt.Sprintf("https://%s/admin", client.config.Domain), client.AdminURL())
}

func TestAssetPath(t *testing.T) {
	client := newTestHTTPClient()
	assert.Equal(t, fmt.Sprintf("%s/assets.json", client.AdminURL()), client.AssetPath(nil))
}

func TestThemesPath(t *testing.T) {
	client := newTestHTTPClient()
	assert.Equal(t, fmt.Sprintf("%s/themes.json", client.AdminURL()), client.ThemesPath())
}

func TestThemePath(t *testing.T) {
	client := newTestHTTPClient()
	assert.Equal(t, fmt.Sprintf("%s/themes/456.json", client.AdminURL()), client.ThemePath(456))
}

func TestAssetList(t *testing.T) {
	server, client := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "fields="+url.QueryEscape(assetDataFields), r.URL.RawQuery)
		fmt.Fprintf(w, jsonFixture("responses/assets"))
	})
	defer server.Close()

	resp, err := client.AssetList()
	assert.Nil(t, err)
	assert.Equal(t, listRequest, resp.Type)
	assert.Equal(t, 2, len(resp.Assets))
}

func TestGetAsset(t *testing.T) {
	server, client := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "asset%5Bkey%5D=file.txt", r.URL.RawQuery)
		fmt.Fprintf(w, jsonFixture("responses/asset"))
	})
	defer server.Close()

	resp, err := client.GetAsset("file.txt")
	assert.Nil(t, err)
	assert.Equal(t, assetRequest, resp.Type)
	assert.Equal(t, "assets/hello.txt", resp.Asset.Key)
}

func TestNewTheme(t *testing.T) {
	server, client := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)

		decoder := json.NewDecoder(r.Body)
		var theme map[string]Theme
		decoder.Decode(&theme)
		defer r.Body.Close()

		assert.Equal(t, Theme{Name: "name", Source: "source", Role: "unpublished"}, theme["theme"])
		fmt.Fprintf(w, jsonFixture("responses/theme"))
	})
	defer server.Close()
	resp, err := client.NewTheme("name", "source")
	assert.Nil(t, err)
	assert.Equal(t, themeRequest, resp.Type)
	assert.Equal(t, "timberland", resp.Theme.Name)
}

func TestGetTheme(t *testing.T) {
	server, client := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		fmt.Fprintf(w, jsonFixture("responses/theme"))
	})
	resp, err := client.GetTheme(123)
	assert.Nil(t, err)
	assert.Equal(t, themeRequest, resp.Type)
	assert.Equal(t, "timberland", resp.Theme.Name)
	server.Close()
}

func TestAssetAction(t *testing.T) {
	server, client := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PUT", r.Method)

		decoder := json.NewDecoder(r.Body)
		var theme map[string]Asset
		decoder.Decode(&theme)
		defer r.Body.Close()

		assert.Equal(t, Asset{Key: "key", Value: "value"}, theme["asset"])
		fmt.Fprintf(w, jsonFixture("responses/asset"))
	})
	defer server.Close()
	resp, err := client.AssetAction(Update, Asset{Key: "key", Value: "value"})
	assert.Nil(t, err)
	assert.Equal(t, assetRequest, resp.Type)
	assert.Equal(t, "assets/hello.txt", resp.Asset.Key)
}

func TestAssetActionStrict(t *testing.T) {
	server, client := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PUT", r.Method)
		assert.Equal(t, r.Header.Get("If-Unmodified-Since"), "version")

		decoder := json.NewDecoder(r.Body)
		var theme map[string]Asset
		decoder.Decode(&theme)
		defer r.Body.Close()

		assert.Equal(t, Asset{Key: "key", Value: "value"}, theme["asset"])
		fmt.Fprintf(w, jsonFixture("responses/asset"))
	})
	defer server.Close()
	resp, err := client.AssetActionStrict(Update, Asset{Key: "key", Value: "value"}, "version")
	assert.Nil(t, err)
	assert.Equal(t, assetRequest, resp.Type)
	assert.Equal(t, "assets/hello.txt", resp.Asset.Key)
}

func TestSendRequest(t *testing.T) {
	server, client := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PUT", r.Method)

		buf := new(bytes.Buffer)
		buf.ReadFrom(r.Body)

		assert.Equal(t, "my string", buf.String())
	})
	defer server.Close()
	req, _ := newShopifyRequest(client.config, assetRequest, Update, server.URL)
	req.setBody(bytes.NewBufferString("my string"))
	resp, err := client.sendRequest(req)
	assert.Nil(t, err)
	assert.Equal(t, assetRequest, resp.Type)
}

func newTestConfig() *Configuration {
	config, _ := NewConfiguration()
	config.Environment = "test"
	config.Domain = "test.myshopify.com"
	config.ThemeID = "123"
	config.Password = "sharknado"
	return config
}

func newTestHTTPClient() *httpClient {
	client, _ := newHTTPClient(newTestConfig())
	return client
}

func newTestServer(handler http.HandlerFunc) (*httptest.Server, *httpClient) {
	client := newTestHTTPClient()
	server := httptest.NewServer(handler)
	client.config.Domain = server.URL
	return server, client
}

func fileFixture(name string) *os.File {
	path := fmt.Sprintf("../fixtures/%s.json", name)
	file, _ := os.Open(path)
	return file
}

func jsonFixture(name string) string {
	bytes, err := ioutil.ReadAll(fileFixture(name))
	if err != nil {
		log.Fatal(err)
	}
	return string(bytes)
}
