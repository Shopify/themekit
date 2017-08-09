package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Shopify/themekit/kit"
	"github.com/Shopify/themekit/kittest"
)

func TestBootstrap(t *testing.T) {
	server := kittest.NewTestServer()
	defer server.Close()
	kittest.Setup()
	defer kittest.Cleanup()
	defer resetArbiter()
	timberFeedPath = server.URL + "/feed"

	assert.NotNil(t, bootstrap(nil, []string{}))

	arbiter.flagConfig.Password = "foo"
	arbiter.flagConfig.Domain = server.URL
	arbiter.flagConfig.Directory = kittest.FixtureProjectPath
	arbiter.setFlagConfig()
	assert.Nil(t, bootstrap(nil, []string{}))

	timberFeedPath = "http://nope.com/nope.json"
	assert.NotNil(t, bootstrap(nil, []string{}))
}

func TestGetNewThemeZipPath(t *testing.T) {
	server := kittest.NewTestServer()
	defer server.Close()
	timberFeedPath = server.URL + "/feed"

	bootstrapURL = "http://github.com/shopify/theme.zip"
	path, err := getNewThemeZipPath()
	assert.Equal(t, bootstrapURL, path)
	assert.Nil(t, err)
	bootstrapURL = ""

	// master just returns early for master version
	bootstrapVersion = "master"
	path, err = getNewThemeZipPath()
	assert.Equal(t, themeZipRoot+"master.zip", path)
	assert.Nil(t, err)

	// valid request
	bootstrapVersion = "v2.0.2"
	path, err = getNewThemeZipPath()
	assert.Equal(t, themeZipRoot+"v2.0.2.zip", path)
	assert.Nil(t, err)

	// not found version
	bootstrapVersion = "vn.0.p"
	path, err = getNewThemeZipPath()
	assert.Equal(t, "", path)
	assert.NotNil(t, err)

	server.Close()

	// server fails to return
	bootstrapVersion = "v2.0.2"
	path, err = getNewThemeZipPath()
	assert.Equal(t, "", path)
	assert.NotNil(t, err)
}

func TestNewGetThemeName(t *testing.T) {
	bootstrapPrefix = "prEfix"
	bootstrapVersion = "4.2.0"
	assert.Equal(t, "prEfixTimber-4.2.0", getNewThemeName())

	bootstrapURL = "http://github.com/shopify/theme.zip"
	assert.Equal(t, "prEfixtheme", getNewThemeName())

	bootstrapName = "bootStrapNaeme"
	assert.Equal(t, "bootStrapNaeme", getNewThemeName())

	bootstrapPrefix = ""
	bootstrapVersion = ""
	bootstrapURL = ""
	bootstrapName = ""
}

func TestDownloadThemeReleaseAtomFeed(t *testing.T) {
	server := kittest.NewTestServer()
	timberFeedPath = server.URL + "/feed"

	feed, err := downloadThemeReleaseAtomFeed()
	assert.Nil(t, err)
	assert.Equal(t, 13, len(feed.Entries))

	timberFeedPath = "http://nope.com/nope.json"
	feed, err = downloadThemeReleaseAtomFeed()
	assert.NotNil(t, err)
	assert.Equal(t, 0, len(feed.Entries))

	server.Close()

	feed, err = downloadThemeReleaseAtomFeed()
	assert.NotNil(t, err)
	assert.Equal(t, 0, len(feed.Entries))
}

func TestFindThemeReleaseWith(t *testing.T) {
	feed := kittest.ReleaseAtom
	entry, err := findThemeReleaseWith(feed, "latest")
	assert.Equal(t, feed.LatestEntry(), entry)
	assert.Nil(t, err)

	entry, err = findThemeReleaseWith(feed, "v2.0.2")
	assert.Equal(t, "v2.0.2", entry.Title)
	assert.Nil(t, err)

	entry, err = findThemeReleaseWith(feed, "nope")
	assert.Equal(t, "Invalid Feed", entry.Title)
	assert.Equal(t, "Invalid Timber Version: nope\nAvailable Versions Are:\n- master\n- latest\n- v2.0.2\n- v2.0.1\n- v2.0.0\n- v1.3.2\n- v1.3.1\n- v1.3.0\n- v1.2.1\n- v1.2.0\n- v1.1.3\n- v1.1.2\n- v1.1.1\n- v1.1.0\n- v1.0.0", err.Error())
	assert.NotNil(t, err)
}

func TestSaveConfiguration(t *testing.T) {
	defer resetArbiter()
	defer kittest.Cleanup()

	kittest.GenerateConfig("example.myshopify.io", true)
	env, _ := kit.LoadEnvironments("config.yml")
	config, _ := env.GetConfiguration(kit.DefaultEnvironment, true)
	assert.Nil(t, saveConfiguration(config))

	kittest.GenerateConfig("example.myshopify.io", false)
	env, _ = kit.LoadEnvironments("config.yml")
	config, _ = env.GetConfiguration(kit.DefaultEnvironment, true)
	assert.NotNil(t, saveConfiguration(config))
}
