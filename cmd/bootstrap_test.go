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

func TestGetZipPath(t *testing.T) {
	server := kittest.NewTestServer()
	defer server.Close()

	bootstrapVersion = "master"
	path, err := getZipPath()
	assert.Equal(t, themeZipRoot+"master.zip", path)
	assert.Nil(t, err)

	bootstrapURL = "http://github.com/shopify/theme.zip"
	path, err = getZipPath()
	assert.Equal(t, bootstrapURL, path)
	assert.Nil(t, err)

	bootstrapURL = ""
	bootstrapVersion = ""
}

func TestGetThemeName(t *testing.T) {
	bootstrapPrefix = "prEfix"
	bootstrapVersion = "4.2.0"
	assert.Equal(t, "prEfixTimber-4.2.0", getThemeName())

	bootstrapURL = "http://github.com/shopify/theme.zip"
	assert.Equal(t, "prEfixtheme", getThemeName())

	bootstrapName = "bootStrapNaeme"
	assert.Equal(t, "bootStrapNaeme", getThemeName())

	bootstrapPrefix = ""
	bootstrapVersion = ""
	bootstrapURL = ""
	bootstrapName = ""
}

func TestZipPath(t *testing.T) {
	assert.Equal(t, themeZipRoot+"foo.zip", zipPath("foo"))
}

func TestZipPathForVersion(t *testing.T) {
	server := kittest.NewTestServer()
	timberFeedPath = server.URL + "/feed"

	// master jst returns early for master version
	path, err := zipPathForVersion("master")
	assert.Equal(t, themeZipRoot+"master.zip", path)
	assert.Nil(t, err)

	// valid request
	path, err = zipPathForVersion("v2.0.2")
	assert.Equal(t, themeZipRoot+"v2.0.2.zip", path)
	assert.Nil(t, err)

	// not found version
	path, err = zipPathForVersion("vn.0.p")
	assert.Equal(t, "", path)
	assert.NotNil(t, err)

	server.Close()

	// server fails to return
	path, err = zipPathForVersion("v2.0.2")
	assert.Equal(t, "", path)
	assert.NotNil(t, err)
}

func TestDownloadAtomFeed(t *testing.T) {
	server := kittest.NewTestServer()
	timberFeedPath = server.URL + "/feed"

	feed, err := downloadAtomFeed()
	assert.Nil(t, err)
	assert.Equal(t, 13, len(feed.Entries))

	timberFeedPath = "http://nope.com/nope.json"
	feed, err = downloadAtomFeed()
	assert.NotNil(t, err)
	assert.Equal(t, 0, len(feed.Entries))

	server.Close()

	feed, err = downloadAtomFeed()
	assert.NotNil(t, err)
	assert.Equal(t, 0, len(feed.Entries))
}

func TestFindReleaseWith(t *testing.T) {
	feed := kittest.ReleaseAtom
	entry, err := findReleaseWith(feed, "latest")
	assert.Equal(t, feed.LatestEntry(), entry)
	assert.Nil(t, err)

	entry, err = findReleaseWith(feed, "v2.0.2")
	assert.Equal(t, "v2.0.2", entry.Title)
	assert.Nil(t, err)

	entry, err = findReleaseWith(feed, "nope")
	assert.Equal(t, "Invalid Feed", entry.Title)
	assert.NotNil(t, err)
}

func TestBuildInvalidVersionError(t *testing.T) {
	feed := kittest.ReleaseAtom
	err := buildInvalidVersionError(feed, "nope")
	assert.Equal(t, "invalid Timber Version: nope\nAvailable Versions Are:\n- master\n- latest\n- v2.0.2\n- v2.0.1\n- v2.0.0\n- v1.3.2\n- v1.3.1\n- v1.3.0\n- v1.2.1\n- v1.2.0\n- v1.1.3\n- v1.1.2\n- v1.1.1\n- v1.1.0\n- v1.0.0", err.Error())
}

func TestSaveConfiguration(t *testing.T) {
	defer resetArbiter()
	defer kittest.Cleanup()

	kittest.GenerateConfig("example.myshopify.io", true)
	env, _ := kit.LoadEnvironments("config.yml")
	config, _ := env.GetConfiguration(kit.DefaultEnvironment)
	assert.Nil(t, saveConfiguration(config))

	kittest.GenerateConfig("example.myshopify.io", false)
	env, _ = kit.LoadEnvironments("config.yml")
	config, _ = env.GetConfiguration(kit.DefaultEnvironment)
	assert.NotNil(t, saveConfiguration(config))
}
