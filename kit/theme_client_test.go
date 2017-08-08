package kit

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Shopify/themekit/kittest"
)

func TestNewThemeClient(t *testing.T) {
	config := &Configuration{}
	client, err := NewThemeClient(config)
	assert.Nil(t, err)
	assert.NotNil(t, client)
	assert.NotNil(t, client.httpClient)
	assert.NotNil(t, client.filter)
	assert.Equal(t, config, client.Config)
	_, err = NewThemeClient(&Configuration{Proxy: "://foo.com"})
	assert.NotNil(t, err)
	_, err = NewThemeClient(&Configuration{Ignores: []string{"nope"}})
	assert.NotNil(t, err)
}

func TestThemeClient_NewFileWatcher(t *testing.T) {
	kittest.GenerateProject()
	defer kittest.Cleanup()
	client, _ := NewThemeClient(&Configuration{Directory: kittest.FixtureProjectPath})
	watcher, err := client.NewFileWatcher("", func(ThemeClient, Asset, EventType) {})
	assert.Nil(t, err)
	assert.NotNil(t, watcher)
}

func TestThemeClient_AssetList(t *testing.T) {
	server := kittest.NewTestServer()
	defer server.Close()
	client, _ := NewThemeClient(&Configuration{Domain: server.URL, ThemeID: "123"})
	assets, err := client.AssetList()
	assert.Nil(t, err)
	assert.Equal(t, 2, len(assets))
	server.Close()
	_, err = client.AssetList()
	assert.NotNil(t, err)
}

func TestThemeClient_Asset(t *testing.T) {
	server := kittest.NewTestServer()
	defer server.Close()
	client, _ := NewThemeClient(&Configuration{Domain: server.URL, ThemeID: "123"})
	asset, err := client.Asset("file.txt")
	assert.Nil(t, err)
	assert.Equal(t, "assets/hello.txt", asset.Key)
	server.Close()
	_, err = client.Asset("file.txt")
	assert.NotNil(t, err)
}

func TestThemeClient_AssetInfo(t *testing.T) {
	server := kittest.NewTestServer()
	defer server.Close()
	client, _ := NewThemeClient(&Configuration{Domain: server.URL, ThemeID: "123"})
	asset, err := client.AssetInfo("file.txt")
	assert.Nil(t, err)
	assert.Equal(t, "assets/hello.txt", asset.Key)
	server.Close()
	_, err = client.AssetInfo("file.txt")
	assert.NotNil(t, err)
}

func TestThemeClient_LocalAssets(t *testing.T) {
	kittest.GenerateProject()
	defer kittest.Cleanup()
	client, _ := NewThemeClient(&Configuration{Directory: kittest.FixtureProjectPath})
	assets, err := client.LocalAssets()
	assert.Nil(t, err)
	assert.Equal(t, 7, len(assets))
	client.Config.Directory = "./nope"
	_, err = client.LocalAssets()
	assert.NotNil(t, err)

	client, _ = NewThemeClient(&Configuration{Directory: kittest.FixtureProjectPath})
	assets, err = client.LocalAssets("assets", "config/settings_data.json")
	assert.Nil(t, err)
	assert.Equal(t, 3, len(assets))

	client, _ = NewThemeClient(&Configuration{Directory: kittest.FixtureProjectPath})
	asset, err := client.LocalAsset("assets/application.js")
	assert.Nil(t, err)
	assert.Equal(t, "assets/application.js", asset.Key)
	_, err = client.LocalAsset("snippets/npe.txt")
	assert.NotNil(t, err)
}

func TestThemeClient_CreateTheme(t *testing.T) {
	server := kittest.NewTestServer()
	defer server.Close()
	flagConfig = Configuration{Domain: server.URL, ThemeID: "123"}
	defer resetConfig()

	client, theme, err := CreateTheme("name", "source")
	if assert.NotNil(t, err) {
		assert.Equal(t, "Invalid options: missing password", err.Error())
	}

	flagConfig = Configuration{Domain: server.URL, ThemeID: "123", Ignores: []string{"nope"}}
	client, theme, err = CreateTheme("name", "source")
	assert.NotNil(t, err)

	flagConfig = Configuration{Domain: server.URL, ThemeID: "123", Password: "test", Proxy: "://foo.com"}
	client, theme, err = CreateTheme("name", "source")
	assert.NotNil(t, err)

	flagConfig = Configuration{Domain: server.URL, ThemeID: "123", Password: "test"}
	client, theme, err = CreateTheme("nope", "source")
	assert.Equal(t, true, strings.Contains(err.Error(), "Cannot create a theme. Last error was"))

	client, theme, err = CreateTheme("name", "source")
	assert.Nil(t, err)
	assert.Equal(t, "timberland", theme.Name)
	assert.Equal(t, fmt.Sprintf("%d", theme.ID), client.Config.ThemeID)
}

func TestThemeClient_CreateAsset(t *testing.T) {
	server := kittest.NewTestServer()
	defer server.Close()
	client, _ := NewThemeClient(&Configuration{Domain: server.URL, ThemeID: "123"})
	resp, err := client.CreateAsset(Asset{Key: "createkey", Value: "value"})
	assert.Nil(t, err)
	assert.Equal(t, "assets/hello.txt", resp.Asset.Key)
	assert.Equal(t, 1, len(server.Requests))
	assert.Equal(t, "PUT", server.Requests[0].Method)
}

func TestThemeClient_UpdateAsset(t *testing.T) {
	server := kittest.NewTestServer()
	defer server.Close()
	client, _ := NewThemeClient(&Configuration{Domain: server.URL, ThemeID: "123"})
	resp, err := client.UpdateAsset(Asset{Key: "updatekey", Value: "value"})
	assert.Nil(t, err)
	assert.Equal(t, "assets/hello.txt", resp.Asset.Key)
	assert.Equal(t, 1, len(server.Requests))
	assert.Equal(t, "PUT", server.Requests[0].Method)
}

func TestThemeClient_DeleteAsset(t *testing.T) {
	server := kittest.NewTestServer()
	defer server.Close()
	client, _ := NewThemeClient(&Configuration{Domain: server.URL, ThemeID: "123"})
	resp, err := client.DeleteAsset(Asset{Key: "deletekey", Value: "value"})
	assert.Nil(t, err)
	assert.Equal(t, "assets/hello.txt", resp.Asset.Key)
	assert.Equal(t, 1, len(server.Requests))
	assert.Equal(t, "DELETE", server.Requests[0].Method)
}

func TestThemeClient_Perform(t *testing.T) {
	server := kittest.NewTestServer()
	defer server.Close()
	client, _ := NewThemeClient(&Configuration{Domain: server.URL, ThemeID: "123"})
	resp, err := client.Perform(Asset{Key: "fookey", Value: "value"}, Create)
	assert.Nil(t, err)
	assert.Equal(t, "assets/hello.txt", resp.Asset.Key)
	assert.Equal(t, 1, len(server.Requests))
	assert.Equal(t, "POST", server.Requests[0].Method)
}

func TestThemeClient_PerformStrict(t *testing.T) {
	server := kittest.NewTestServer()
	defer server.Close()
	client, _ := NewThemeClient(&Configuration{Domain: server.URL, ThemeID: "123"})
	resp, err := client.PerformStrict(Asset{Key: "fookey", Value: "value"}, Create, "version")
	assert.Nil(t, err)
	assert.Equal(t, "assets/hello.txt", resp.Asset.Key)
	assert.Equal(t, 1, len(server.Requests))
	assert.Equal(t, "POST", server.Requests[0].Method)
	assert.Equal(t, server.Requests[0].Header.Get("If-Unmodified-Since"), "version")
}

func TestThemeClient_AfterHooks(t *testing.T) {
	server := kittest.NewTestServer()
	defer server.Close()
	client, _ := NewThemeClient(&Configuration{Domain: server.URL, ThemeID: "123"})
	_, err := client.Perform(Asset{Key: "templates/template.html", Value: "value();"}, Update)
	assert.Nil(t, err)
	assert.Equal(t, 3, len(server.Requests))
	assert.Equal(t, "PUT", server.Requests[0].Method)
	assert.Equal(t, "DELETE", server.Requests[1].Method)
	assert.Equal(t, "PUT", server.Requests[2].Method)
}
