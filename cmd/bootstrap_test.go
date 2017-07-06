package cmd

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Shopify/themekit/kit"
)

func init() {
	resetArbiter()
}

func TestBootstrap(t *testing.T) {
	server := newTestServer()
	defer server.Close()
	defer resetArbiter()
	defer os.Remove("config.yml")

	err := bootstrap(nil, []string{})
	assert.NotNil(t, err)

	arbiter.flagConfig.Password = "foo"
	arbiter.flagConfig.Domain = server.URL + "/domain"
	arbiter.setFlagConfig()

	err = bootstrap(nil, []string{})
	assert.Nil(t, err)

	timberFeedPath = "http://nope.com/nope.json"
	err = bootstrap(nil, []string{})
	assert.NotNil(t, err)
}

func TestGetZipPath(t *testing.T) {
	server := newTestServer()
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
	server := newTestServer()

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
	server := newTestServer()

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
	feed := loadAtom()
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
	feed := loadAtom()
	err := buildInvalidVersionError(feed, "nope")
	assert.Equal(t, "invalid Timber Version: nope\nAvailable Versions Are:\n- master\n- latest\n- v2.0.2\n- v2.0.1\n- v2.0.0\n- v1.3.2\n- v1.3.1\n- v1.3.0\n- v1.2.1\n- v1.2.0\n- v1.1.3\n- v1.1.2\n- v1.1.1\n- v1.1.0\n- v1.0.0", err.Error())
}

func TestSaveConfiguration(t *testing.T) {
	defer os.Remove("config.yml")
	defer resetArbiter()

	arbiter.configPath = goodEnvirontmentPath
	env, err := kit.LoadEnvironments(arbiter.configPath)
	config, _ := env.GetConfiguration(kit.DefaultEnvironment)

	err = saveConfiguration(config)
	assert.Nil(t, err)

	arbiter.configPath = badEnvirontmentPath
	err = saveConfiguration(config)
	assert.NotNil(t, err)

	arbiter.configPath = "config.yml"
	err = saveConfiguration(config)
	assert.Nil(t, err)
}
