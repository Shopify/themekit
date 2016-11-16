package kit

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"runtime"
	"testing"

	"github.com/hashicorp/go-version"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

var (
	updatePath      = clean("../fixtures/updateme")
	oldFile         = []byte{0xDE, 0xAD, 0xBE, 0xEF}
	newFile         = []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06}
	newFileChecksum = md5.Sum(newFile)
)

func TestLibraryInfo(t *testing.T) {
	messageSeparator := "\n----------------------------------------------------------------\n"
	info := fmt.Sprintf("\t%s %s", "ThemeKit - Shopify Theme Utilities", ThemeKitVersion.String())
	assert.Equal(t, fmt.Sprintf("%s%s%s", messageSeparator, info, messageSeparator), LibraryInfo())
}

type VersionTestSuite struct {
	suite.Suite
}

func (suite *VersionTestSuite) SetupSuite() {
	ThemeKitVersion, _ = version.NewVersion("0.5.0")
}

func (suite *VersionTestSuite) SetupTest() {
	file, err := os.Create(updatePath)
	if err == nil {
		file.Close()
	}
}

func (suite *VersionTestSuite) TearDownTest() {
	os.Remove(updatePath)
}

func (suite *VersionTestSuite) TestIsNewUpdateAvailable() {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, jsonFixture("responses/all_releases"))
	}))
	releasesURL = server.URL
	defer server.Close()
	ThemeKitVersion, _ = version.NewVersion("20.0.0")
	assert.Equal(suite.T(), false, IsNewUpdateAvailable())

	ThemeKitVersion, _ = version.NewVersion("0.0.0")
	assert.Equal(suite.T(), true, IsNewUpdateAvailable())
}

func (suite *VersionTestSuite) TestInstallThemeKitVersion() {
	requests := 0
	var server *httptest.Server
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if requests <= 1 {
			fmt.Fprintf(w, jsonFixture("responses/all_releases"))
		} else if requests == 2 {
			out, _ := json.Marshal([]release{
				{
					Version: "20.0.0",
					Platforms: []platform{
						{
							Name:       runtime.GOOS + "-" + runtime.GOARCH,
							URL:        server.URL,
							Digest:     hex.EncodeToString(newFileChecksum[:]),
							TargetPath: updatePath,
						},
					},
				},
			})

			fmt.Fprintf(w, string(out))
		} else {
			fmt.Fprintf(w, string(newFile))
		}
		requests++
	}))
	defer server.Close()
	releasesURL = server.URL

	ThemeKitVersion, _ = version.NewVersion("0.4.7")
	err := InstallThemeKitVersion("latest")
	assert.Equal(suite.T(), "no applicable update available", err.Error())

	ThemeKitVersion, _ = version.NewVersion("0.4.4")
	err = InstallThemeKitVersion("0.0.0")
	assert.Equal(suite.T(), "version 0.0.0 not found", err.Error())

	ThemeKitVersion, _ = version.NewVersion("0.4.4")
	err = InstallThemeKitVersion("latest")
	assert.Nil(suite.T(), err)
}

func (suite *VersionTestSuite) TestFetchReleases() {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, jsonFixture("responses/all_releases"))
	}))
	releasesURL = server.URL

	releases, err := fetchReleases()
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), 4, len(releases))
	server.Close()

	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "this is not json")
	}))
	releasesURL = server.URL
	_, err = fetchReleases()
	assert.NotNil(suite.T(), err)
	server.Close()

	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
		fmt.Fprintf(w, "404")
	}))
	releasesURL = server.URL
	_, err = fetchReleases()
	assert.NotNil(suite.T(), err)
	server.Close()
}

func (suite *VersionTestSuite) TestApplyUpdate() {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, string(newFile))
	}))

	err := applyUpdate(platform{
		URL:        server.URL,
		Digest:     hex.EncodeToString(newFileChecksum[:]),
		TargetPath: updatePath,
	})
	assert.Nil(suite.T(), err)

	buf, err := ioutil.ReadFile(updatePath)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), newFile, buf)
	server.Close()

	err = applyUpdate(platform{})
	assert.NotNil(suite.T(), err)

	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
		fmt.Fprintf(w, "404")
	}))
	err = applyUpdate(platform{URL: server.URL})
	assert.NotNil(suite.T(), err)
	server.Close()
}

func TestVersionTestSuite(t *testing.T) {
	suite.Run(t, new(VersionTestSuite))
}
