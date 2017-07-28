package kit

import (
	"encoding/hex"
	"io/ioutil"
	"testing"

	"github.com/hashicorp/go-version"
	"github.com/stretchr/testify/assert"

	"github.com/Shopify/themekit/kittest"
)

func TestIsNewUpdateAvailable(t *testing.T) {
	kittest.Setup()
	defer kittest.Cleanup()
	server := kittest.NewTestServer()
	defer server.Close()
	ThemeKitLatestURL = server.URL + "/themekit_latest"
	ThemeKitVersion, _ = version.NewVersion("20.0.0")
	assert.False(t, IsNewUpdateAvailable())
	ThemeKitVersion, _ = version.NewVersion("0.0.0")
	assert.True(t, IsNewUpdateAvailable())
	server.Close()
	assert.False(t, IsNewUpdateAvailable())
}

func TestInstallThemeKitVersion(t *testing.T) {
	kittest.Setup()
	defer kittest.Cleanup()
	server := kittest.NewTestServer()
	defer server.Close()
	ThemeKitReleasesURL = server.URL + "/themekit_update"
	ThemeKitLatestURL = server.URL + "/themekit_latest"
	ThemeKitVersion, _ = version.NewVersion("0.4.7")
	err := InstallThemeKitVersion("latest")
	assert.Equal(t, "no applicable update available", err.Error())
	ThemeKitReleasesURL = server.URL + "/themekit_system_update"
	ThemeKitLatestURL = server.URL + "/themekit_latest_system_update"
	ThemeKitVersion, _ = version.NewVersion("0.4.4")
	err = InstallThemeKitVersion("0.0.0")
	assert.Equal(t, "version 0.0.0 not found", err.Error())
	assert.Nil(t, InstallThemeKitVersion("latest"))
	server.Close()
	assert.NotNil(t, InstallThemeKitVersion("latest"))
	assert.NotNil(t, InstallThemeKitVersion("0.4.7"))
}

func TestFetchReleases(t *testing.T) {
	server := kittest.NewTestServer()
	defer server.Close()
	ThemeKitReleasesURL = server.URL + "/themekit_update"
	releases, err := FetchReleases()
	assert.Nil(t, err)
	assert.Equal(t, 4, len(releases))
	ThemeKitReleasesURL = server.URL + "/not_json"
	_, err = FetchReleases()
	assert.NotNil(t, err)
	ThemeKitReleasesURL = server.URL + "/doesntexist"
	_, err = FetchReleases()
	assert.NotNil(t, err)
	server.Close()
	_, err = FetchReleases()
	assert.NotNil(t, err)
}

func TestApplyUpdate(t *testing.T) {
	kittest.Setup()
	defer kittest.Cleanup()
	server := kittest.NewTestServer()
	defer server.Close()
	assert.Nil(t, applyUpdate(Platform{
		URL:        server.URL + "/release_download",
		Digest:     hex.EncodeToString(kittest.NewUpdateFileChecksum[:]),
		TargetPath: kittest.UpdateFilePath,
	}))
	buf, err := ioutil.ReadFile(kittest.UpdateFilePath)
	assert.Nil(t, err)
	assert.Equal(t, kittest.NewUpdateFile, buf)
	assert.NotNil(t, applyUpdate(Platform{}))
	assert.NotNil(t, applyUpdate(Platform{Digest: "abcde"}))
	assert.NotNil(t, applyUpdate(Platform{URL: server.URL + "/doesntexist"}))
}
