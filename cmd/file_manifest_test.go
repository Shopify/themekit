package cmd

import (
	"os"
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
	assert.Equal(t, store.store, s)
}

func TestGenerateRemote(t *testing.T) {
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

func TestParseTime(t *testing.T) {
	time := parseTime("2012-07-06T02:04:21-11:00")
	assert.Equal(t, time.Hour(), 2)
	assert.Equal(t, time.Minute(), 4)
	assert.Equal(t, time.Year(), 2012)
	assert.Equal(t, time.Day(), 6)
	assert.True(t, parseTime("").IsZero())
}

func TestDiffDates(t *testing.T) {
	manifest := &fileManifest{
		local:  make(map[string]map[string]string),
		remote: make(map[string]map[string]string),
	}

	local, remote := manifest.diffDates("asset.js", "production", "development")
	assert.True(t, local.IsZero())
	assert.True(t, remote.IsZero())

	manifest.local["asset.js"] = map[string]string{"development": "2012-07-06T02:04:21-11:00"}
	manifest.remote["asset.js"] = map[string]string{"production": "2012-07-06T02:04:21-11:00"}
	local, remote = manifest.diffDates("asset.js", "production", "development")
	assert.False(t, local.IsZero())
	assert.False(t, remote.IsZero())
}

func TestShould(t *testing.T) {
	file, env, now, then := "asset.js", "development", "2017-07-06T02:04:21-11:00", "2012-07-06T02:04:21-11:00"
	manifest := &fileManifest{
		local:  map[string]map[string]string{file: {env: then}},
		remote: map[string]map[string]string{file: {env: then}},
	}

	assert.False(t, manifest.Should(kit.Update, file, env))
	assert.False(t, manifest.Should(kit.Remove, file, env))
	assert.False(t, manifest.Should(kit.Retrieve, file, env))

	manifest.local[file][env] = now
	assert.True(t, manifest.Should(kit.Update, file, env))
	assert.True(t, manifest.Should(kit.Remove, file, env))
	assert.False(t, manifest.Should(kit.Retrieve, file, env))

	manifest.local[file][env] = then
	manifest.remote[file][env] = now
	assert.False(t, manifest.Should(kit.Update, file, env))
	assert.False(t, manifest.Should(kit.Remove, file, env))
	assert.True(t, manifest.Should(kit.Retrieve, file, env))

	delete(manifest.remote[file], env)
	assert.True(t, manifest.Should(kit.Update, file, env))

	assert.False(t, manifest.Should(kit.EventType(25), file, env))
}

func TestFetchableFiles(t *testing.T) {
	server := kittest.NewTestServer()
	defer server.Close()
	assert.Nil(t, kittest.GenerateConfig(server.URL, true))
	defer kittest.Cleanup()

	env, now := "development", "2017-07-06T02:04:21-11:00"
	manifest := &fileManifest{
		remote: map[string]map[string]string{
			"assets/goodbye.txt": {env: now},
			"assets/hello.txt":   {env: now},
		},
	}

	filenames := manifest.FetchableFiles([]string{"assets/hello.txt"}, env)
	assert.Equal(t, len(filenames), 1)

	filenames = manifest.FetchableFiles([]string{"assets/goodbye*", "templates/404.html"}, env)
	assert.Equal(t, filenames, []string{"templates/404.html", "assets/goodbye.txt"})
}

func TestDiff(t *testing.T) {
	dstenv, srcenv, now, then := "production", "development", "2017-07-06T02:04:21-11:00", "2012-07-06T02:04:21-11:00"
	manifest := &fileManifest{
		local: map[string]map[string]string{
			"asset1.js": {srcenv: now},
			"asset2.js": {srcenv: then},
		},
		remote: map[string]map[string]string{
			"asset2.js": {dstenv: now},
			"asset3.js": {dstenv: then},
		},
	}

	actions := map[string]assetAction{
		"asset1.js": {event: kit.Update},
		"asset2.js": {event: kit.Update},
		"asset3.js": {event: kit.Remove},
	}

	diff := manifest.Diff(actions, dstenv, srcenv)
	assert.Equal(t, 1, len(diff.Created))
	assert.Equal(t, 1, len(diff.Updated))
	assert.Equal(t, 1, len(diff.Removed))
}

func TestManifestSet(t *testing.T) {
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
		assert.NotNil(t, err)
		assert.Nil(t, manifest.Set("test", "development", "test"))
		_, err = os.Stat(storeName)
		assert.Nil(t, err)
	}
}

func TestDelete(t *testing.T) {
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
		assert.Nil(t, manifest.Set("test.txt", "development", "123456"))
		assert.Nil(t, manifest.Delete("test.txt", "development"))
	}
}

func TestGet(t *testing.T) {
	server := kittest.NewTestServer()
	defer server.Close()
	assert.Nil(t, kittest.GenerateConfig(server.URL, true))
	defer kittest.Cleanup()
	defer resetArbiter()

	client, err := getClient()
	if assert.Nil(t, err) {
		manifest, err := newFileManifest("", []kit.ThemeClient{client})
		assert.Nil(t, err)
		_, err = manifest.Get("test.txt", "development")
		assert.Nil(t, err)
		_, err = manifest.Get("", "development")
		assert.NotNil(t, err)
		assert.Nil(t, manifest.Set("test.txt", "development", "123456"))
		version, err := manifest.Get("test.txt", "development")
		assert.Nil(t, err)
		assert.Equal(t, version, "123456")
	}
}
