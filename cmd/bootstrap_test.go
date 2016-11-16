package cmd

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/Shopify/themekit/cmd/internal/atom"
)

type BootstrapTestSuite struct {
	suite.Suite
}

func (suite *BootstrapTestSuite) TestBootstrap() {
	responses := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println(r.URL.Path)
		if r.URL.Path == "/feed" {
			file, _ := os.Open("../fixtures/releases.atom")
			bytes, _ := ioutil.ReadAll(file)
			fmt.Fprintf(w, string(bytes))
		} else if r.URL.Path == "/domain/admin/themes.json" {
			fmt.Fprintf(w, jsonFixture("responses/theme"))
		} else if r.URL.Path == "/domain/admin/themes/0.json" {
			fmt.Fprintf(w, jsonFixture("responses/assets"))
		}
		responses++
	}))
	defer server.Close()
	timberFeedPath = server.URL + "/feed"
	themeZipRoot = server.URL + "/zip"

	err := bootstrap()
	assert.NotNil(suite.T(), err)

	flagConfig.Directory = "../fixtures/bootstrap"
	flagConfig.Password = "foo"
	flagConfig.Domain = server.URL + "/domain"
	setFlagConfig()
	err = bootstrap()
	fmt.Println(err)
	assert.Nil(suite.T(), err)

	flagConfig.Directory = ""
	flagConfig.Password = ""
	flagConfig.Domain = ""
	setFlagConfig()

	os.Remove("./config.yml")
}

func (suite *BootstrapTestSuite) TestZipPath() {
	assert.Equal(suite.T(), themeZipRoot+"foo.zip", zipPath("foo"))
}

func (suite *BootstrapTestSuite) TestZipPathForVersion() {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		file, _ := os.Open("../fixtures/releases.atom")
		bytes, _ := ioutil.ReadAll(file)
		fmt.Fprintf(w, string(bytes))
	}))
	timberFeedPath = server.URL

	path, err := zipPathForVersion("master")
	assert.Equal(suite.T(), themeZipRoot+"master.zip", path)
	assert.Nil(suite.T(), err)

	path, err = zipPathForVersion("v2.0.2")
	assert.Equal(suite.T(), themeZipRoot+"v2.0.2.zip", path)
	assert.Nil(suite.T(), err)

	path, err = zipPathForVersion("vn.0.p")
	assert.Equal(suite.T(), "", path)
	assert.NotNil(suite.T(), err)

	server.Close()

	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
		fmt.Fprintf(w, "404")
	}))
	timberFeedPath = server.URL

	path, err = zipPathForVersion("v2.0.2")
	assert.Equal(suite.T(), "", path)
	assert.NotNil(suite.T(), err)
	server.Close()
}

func (suite *BootstrapTestSuite) TestDownloadAtomFeed() {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		file, _ := os.Open("../fixtures/releases.atom")
		bytes, _ := ioutil.ReadAll(file)
		fmt.Fprintf(w, string(bytes))
	}))
	timberFeedPath = server.URL

	feed, err := downloadAtomFeed()
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), 13, len(feed.Entries))
	server.Close()

	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "not atom")
	}))
	timberFeedPath = server.URL

	feed, err = downloadAtomFeed()
	assert.NotNil(suite.T(), err)
	assert.Equal(suite.T(), 0, len(feed.Entries))
	server.Close()

	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
		fmt.Fprintf(w, "404")
	}))
	timberFeedPath = server.URL

	feed, err = downloadAtomFeed()
	assert.NotNil(suite.T(), err)
	assert.Equal(suite.T(), 0, len(feed.Entries))
	server.Close()
}

func (suite *BootstrapTestSuite) TestFindReleaseWith() {
	feed := loadAtom()
	entry, err := findReleaseWith(feed, "latest")
	assert.Equal(suite.T(), feed.LatestEntry(), entry)
	assert.Nil(suite.T(), err)

	entry, err = findReleaseWith(feed, "v2.0.2")
	assert.Equal(suite.T(), "v2.0.2", entry.Title)
	assert.Nil(suite.T(), err)

	entry, err = findReleaseWith(feed, "nope")
	assert.Equal(suite.T(), "Invalid Feed", entry.Title)
	assert.NotNil(suite.T(), err)
}

func (suite *BootstrapTestSuite) TestBuildInvalidVersionError() {
	feed := loadAtom()
	err := buildInvalidVersionError(feed, "nope")
	assert.Equal(suite.T(), "invalid Timber Version: nope\nAvailable Versions Are:\n- master\n- latest\n- v2.0.2\n- v2.0.1\n- v2.0.0\n- v1.3.2\n- v1.3.1\n- v1.3.0\n- v1.2.1\n- v1.2.0\n- v1.1.3\n- v1.1.2\n- v1.1.1\n- v1.1.0\n- v1.0.0", err.Error())
}

func TestBootstrapTestSuite(t *testing.T) {
	suite.Run(t, new(BootstrapTestSuite))
}

func loadAtom() atom.Feed {
	stream, _ := os.Open("../fixtures/releases.atom")
	feed, _ := atom.LoadFeed(stream)
	return feed
}
