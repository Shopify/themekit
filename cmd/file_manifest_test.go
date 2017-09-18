package cmd

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Shopify/themekit/cmd/ystore"
	"github.com/Shopify/themekit/kit"
	"github.com/Shopify/themekit/kittest"
)

func TestNewFileManifest(t *testing.T) {
	store, err := newFileManifest("", []kit.ThemeClient{})
	assert.Nil(t, err)
	s, _ := ystore.New(storeName)
	defer s.Drop()
	assert.Equal(t, store.store, s)
}

func TestFileManifest_GenerateRemote(t *testing.T) {
	server := kittest.NewTestServer()
	defer server.Close()
	assert.Nil(t, kittest.GenerateConfig(server.URL, true))
	defer kittest.Cleanup()
	defer resetArbiter()

	manifest, err := newFileManifest("", []kit.ThemeClient{})
	assert.Nil(t, err)

	client, err := getClient()
	if assert.Nil(t, err) {
		assert.Nil(t, manifest.generateRemote([]kit.ThemeClient{client}))
		assert.Equal(t, 2, len(manifest.remote))

		server.Close()
		assert.NotNil(t, manifest.generateRemote([]kit.ThemeClient{client}))
	}
}

func TestFileManifest_Prune(t *testing.T) {
	server := kittest.NewTestServer()
	defer server.Close()
	assert.Nil(t, kittest.GenerateConfig(server.URL, true))
	defer kittest.Cleanup()
	defer resetArbiter()
	assert.Nil(t, arbiter.generateThemeClients(nil, []string{}))

	file, env, now := "asset.js", "development", "2017-07-06T02:04:21-11:00"
	store, _ := ystore.New(storeName)
	defer store.Drop()
	manifest := &fileManifest{
		store:  store,
		local:  map[string]map[string]string{file: {env: now, "other": now}},
		remote: map[string]map[string]string{},
	}
	kittest.TouchFixtureFile("asset.js", "")
	assert.Equal(t, ystore.ErrorCollectionNotFound, manifest.prune(arbiter.activeThemeClients))

	store.Write(file, "other", now)
	assert.Nil(t, manifest.prune(arbiter.activeThemeClients))

	_, ok := manifest.local[file]["other"]
	assert.False(t, ok)

	manifest.local = map[string]map[string]string{"": {env: now}}
	assert.NotNil(t, manifest.prune(arbiter.activeThemeClients))
}

func TestParseTime(t *testing.T) {
	time := parseTime("2012-07-06T02:04:21-11:00")
	assert.Equal(t, time.Hour(), 2)
	assert.Equal(t, time.Minute(), 4)
	assert.Equal(t, time.Year(), 2012)
	assert.Equal(t, time.Day(), 6)
	assert.True(t, parseTime("").IsZero())
}

func TestFileManifest_DiffDates(t *testing.T) {
	manifest := &fileManifest{
		local:  make(map[string]map[string]string),
		remote: make(map[string]map[string]string),
	}

	local, remote := manifest.diffDates("asset.js", "production")
	assert.True(t, local.IsZero())
	assert.True(t, remote.IsZero())

	manifest.local["asset.js"] = map[string]string{"production": "2012-07-06T02:04:21-11:00"}
	manifest.remote["asset.js"] = map[string]string{"production": "2012-07-06T02:04:21-11:00"}
	local, remote = manifest.diffDates("asset.js", "production")
	assert.False(t, local.IsZero())
	assert.False(t, remote.IsZero())
}

func TestFileManifest_Should(t *testing.T) {
	asset := kit.Asset{
		Key:   "asset.js",
		Value: "// this is content",
	}
	checksum, _ := asset.CheckSum()
	println(checksum)

	env, now, then := "development", "2017-07-06T02:04:21-11:00", "2012-07-06T02:04:21-11:00"
	manifest := &fileManifest{
		local:  map[string]map[string]string{asset.Key: {env: strings.Join([]string{then, "e22eb25ed76d48248d849e3107754642"}, versionSeparator)}},
		remote: map[string]map[string]string{asset.Key: {env: then}},
	}

	assert.False(t, manifest.Should(kit.Update, asset, env))
	assert.True(t, manifest.Should(kit.Remove, asset, env))
	assert.False(t, manifest.Should(kit.Retrieve, asset, env))

	manifest.local[asset.Key][env] = now
	assert.True(t, manifest.Should(kit.Update, asset, env))
	assert.True(t, manifest.Should(kit.Remove, asset, env))
	assert.False(t, manifest.Should(kit.Retrieve, asset, env))

	manifest.local[asset.Key][env] = strings.Join([]string{then, "e22eb25ed76d48248d849e3107754642"}, versionSeparator)
	manifest.remote[asset.Key][env] = now
	assert.False(t, manifest.Should(kit.Update, asset, env))
	assert.False(t, manifest.Should(kit.Remove, asset, env))
	assert.True(t, manifest.Should(kit.Retrieve, asset, env))

	manifest.local[asset.Key][env] = strings.Join([]string{then, "nope"}, versionSeparator)
	assert.True(t, manifest.Should(kit.Update, asset, env))

	delete(manifest.remote[asset.Key], env)
	assert.True(t, manifest.Should(kit.Update, asset, env))
	assert.True(t, manifest.Should(kit.Remove, asset, env))
	assert.False(t, manifest.Should(kit.Retrieve, asset, env))

	delete(manifest.local[asset.Key], env)
	assert.True(t, manifest.Should(kit.Update, asset, env))
	assert.True(t, manifest.Should(kit.Remove, asset, env))
	assert.True(t, manifest.Should(kit.Retrieve, asset, env))

	assert.False(t, manifest.Should(kit.EventType(25), asset, env))
}

func TestFileManifest_FetchableFiles(t *testing.T) {
	server := kittest.NewTestServer()
	defer server.Close()
	assert.Nil(t, kittest.GenerateConfig(server.URL, true))
	defer kittest.Cleanup()

	env, now := "development", "2017-07-06T02:04:21-11:00"
	manifest := &fileManifest{
		remote: map[string]map[string]string{
			"assets/goodbye.txt": {env: now},
			"assets/hello.txt":   {env: now},
			"assets/no.txt":      {"other": now},
		},
	}

	filenames := manifest.FetchableFiles([]string{}, env)
	assert.Equal(t, len(filenames), 2)

	filenames = manifest.FetchableFiles([]string{"assets/hello.txt"}, env)
	assert.Equal(t, len(filenames), 1)

	filenames = manifest.FetchableFiles([]string{"assets/goodbye*", "templates/404.html"}, env)
	assert.Equal(t, filenames, []string{"templates/404.html", "assets/goodbye.txt"})
}

func TestFileManifest_Diff(t *testing.T) {
	env, now, then := "production", "2017-07-06T02:04:21-11:00", "2012-07-06T02:04:21-11:00"
	manifest := &fileManifest{
		local: map[string]map[string]string{
			"asset1.js": {env: now},
			"asset2.js": {env: then},
		},
		remote: map[string]map[string]string{
			"asset2.js": {env: now},
			"asset3.js": {env: then},
		},
	}

	actions := map[string]assetAction{
		"asset1.js": {event: kit.Update},
		"asset2.js": {event: kit.Update},
		"asset3.js": {event: kit.Remove},
	}

	diff := manifest.Diff(actions, env)
	assert.Equal(t, 1, len(diff.Created))
	assert.Equal(t, 1, len(diff.Updated))
	assert.Equal(t, 1, len(diff.Removed))
}

func TestFileManifest_Set(t *testing.T) {
	server := kittest.NewTestServer()
	defer server.Close()
	assert.Nil(t, kittest.GenerateConfig(server.URL, true))
	defer kittest.Cleanup()
	defer resetArbiter()

	client, err := getClient()
	if assert.Nil(t, err) {
		manifest, err := newFileManifest("", []kit.ThemeClient{client})
		assert.Nil(t, err)

		_, err = os.Stat(storeName)
		assert.Nil(t, err)
		assert.Nil(t, manifest.Set("test", "development", "test", "akjblksb1242ljhbl243"))
		_, err = os.Stat(storeName)
		assert.Nil(t, err)
	}
}

func TestFileManifest_Delete(t *testing.T) {
	server := kittest.NewTestServer()
	defer server.Close()
	assert.Nil(t, kittest.GenerateConfig(server.URL, true))
	defer kittest.Cleanup()
	defer resetArbiter()

	client, err := getClient()
	if assert.Nil(t, err) {
		manifest, err := newFileManifest("", []kit.ThemeClient{client})
		assert.Nil(t, err)
		assert.NotNil(t, manifest.Delete("test.txt", "development"))
		assert.Nil(t, manifest.Set("test.txt", "development", "123456", "123456"))
		assert.Nil(t, manifest.Delete("test.txt", "development"))
	}
}

func TestFileManifest_Get(t *testing.T) {
	server := kittest.NewTestServer()
	defer server.Close()
	assert.Nil(t, kittest.GenerateConfig(server.URL, true))
	defer kittest.Cleanup()
	defer resetArbiter()

	client, err := getClient()
	if assert.Nil(t, err) {
		manifest, err := newFileManifest("", []kit.ThemeClient{client})
		assert.Nil(t, err)
		_, _, err = manifest.Get("test.txt", "development")
		assert.Nil(t, err)
		_, _, err = manifest.Get("", "development")
		assert.NotNil(t, err)
		assert.Nil(t, manifest.Set("test.txt", "development", "123456", "123sum456"))
		version, sum, err := manifest.Get("test.txt", "development")
		assert.Nil(t, err)
		assert.Equal(t, version, "123456")
		assert.Equal(t, sum, "123sum456")
	}
}
