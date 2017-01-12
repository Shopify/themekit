package kit

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sort"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type ThemeClientTestSuite struct {
	suite.Suite
	config *Configuration
	client ThemeClient
}

func (suite *ThemeClientTestSuite) SetupTest() {
	suite.config, _ = NewConfiguration()
	suite.config.Domain = "test.myshopify.com"
	suite.config.ThemeID = "123"
	suite.config.Password = "sharknado"
	suite.config.Directory = "../fixtures/project"
	suite.config.IgnoredFiles = []string{"fookeybee"}
	suite.client, _ = NewThemeClient(suite.config)
}

func (suite *ThemeClientTestSuite) TestNewThemeClient() {
	client, err := NewThemeClient(suite.config)
	assert.Nil(suite.T(), err)
	assert.NotNil(suite.T(), client)
	assert.NotNil(suite.T(), client.httpClient)
	assert.NotNil(suite.T(), client.filter)
	assert.Equal(suite.T(), suite.config, client.Config)
}

func (suite *ThemeClientTestSuite) TestNewThemeClientError() {
	suite.config.Proxy = "://foo.com"
	_, err := NewThemeClient(suite.config)
	assert.NotNil(suite.T(), err)

	suite.config.Proxy = ""
	suite.config.Ignores = []string{"nope"}
	_, err = NewThemeClient(suite.config)
	assert.NotNil(suite.T(), err)
}

func (suite *ThemeClientTestSuite) TestNewFileWatcher() {
	watcher, err := suite.client.NewFileWatcher("", func(ThemeClient, Asset, EventType) {})
	assert.Nil(suite.T(), err)
	assert.NotNil(suite.T(), watcher)
}

func (suite *ThemeClientTestSuite) TestAssetList() {
	server := suite.NewTestServer(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(suite.T(), "GET", r.Method)
		assert.Equal(suite.T(), "fields=key,attachment,value", r.URL.RawQuery)
		fmt.Fprintf(w, jsonFixture("responses/assets_raw"))
	})
	defer server.Close()

	expected := map[string][]Asset{}
	bytes := []byte(jsonFixture("responses/assets_filtered"))
	json.Unmarshal(bytes, &expected)
	sort.Sort(ByAsset(expected["assets"]))

	assets, err := suite.client.AssetList()
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), expected["assets"], assets)
}

func (suite *ThemeClientTestSuite) TestAsset() {
	server := suite.NewTestServer(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(suite.T(), "GET", r.Method)
		assert.Equal(suite.T(), "fields=key,attachment,value&asset[key]=file.txt", r.URL.RawQuery)
		fmt.Fprintf(w, jsonFixture("responses/asset"))
	})
	defer server.Close()
	asset, err := suite.client.Asset("file.txt")
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), "assets/hello.txt", asset.Key)
}

func (suite *ThemeClientTestSuite) TestLocalAssets() {
	assets, err := suite.client.LocalAssets()
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), 7, len(assets))

	suite.client.Config.Directory = "./nope"
	_, err = suite.client.LocalAssets()
	assert.NotNil(suite.T(), err)
}

func (suite *ThemeClientTestSuite) TestLocalAsset() {
	asset, err := suite.client.LocalAsset("assets/application.js")
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), "assets/application.js", asset.Key)

	_, err = suite.client.LocalAsset("snippets/npe.txt")
	assert.NotNil(suite.T(), err)
}

func (suite *ThemeClientTestSuite) TestCreateTheme() {
	responses := 0
	server := suite.NewTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			fmt.Fprintf(w, jsonFixture("responses/theme"))
		} else if r.Method == "POST" {
			if responses < 3 {
				fmt.Fprintf(w, jsonFixture("responses/theme_error"))
			} else {
				fmt.Fprintf(w, jsonFixture("responses/theme"))
			}
		}
		responses++
	})
	defer server.Close()

	client, theme, err := CreateTheme("name", "source")
	if assert.NotNil(suite.T(), err) {
		assert.Equal(suite.T(), "Invalid options: missing store domain,missing password", err.Error())
	}

	suite.config.Ignores = []string{"nope"}
	SetFlagConfig(*suite.config)
	client, theme, err = CreateTheme("name", "source")
	assert.NotNil(suite.T(), err)

	suite.config.Ignores = []string{}
	suite.config.Domain = server.URL

	SetFlagConfig(*suite.config)
	client, theme, err = CreateTheme("name", "source")
	if assert.NotNil(suite.T(), err) {
		assert.Equal(suite.T(), true, strings.Contains(err.Error(), "Cannot create a theme. Last error was"))
	}

	client, theme, err = CreateTheme("name", "source")
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), "timberland", theme.Name)
	assert.Equal(suite.T(), fmt.Sprintf("%d", theme.ID), client.Config.ThemeID)
}

func (suite *ThemeClientTestSuite) TestCreateAsset() {
	server := suite.NewTestServer(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(suite.T(), "PUT", r.Method)

		decoder := json.NewDecoder(r.Body)
		var t map[string]Asset
		decoder.Decode(&t)
		defer r.Body.Close()

		assert.Equal(suite.T(), Asset{Key: "createkey", Value: "value"}, t["asset"])
		fmt.Fprintf(w, jsonFixture("responses/asset"))
	})
	defer server.Close()

	suite.client.CreateAsset(Asset{Key: "createkey", Value: "value"})
}

func (suite *ThemeClientTestSuite) TestUpdateAsset() {
	server := suite.NewTestServer(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(suite.T(), "PUT", r.Method)

		decoder := json.NewDecoder(r.Body)
		var t map[string]Asset
		decoder.Decode(&t)
		defer r.Body.Close()

		assert.Equal(suite.T(), Asset{Key: "updatekey", Value: "value"}, t["asset"])
		fmt.Fprintf(w, jsonFixture("responses/asset"))
	})
	defer server.Close()

	suite.client.UpdateAsset(Asset{Key: "updatekey", Value: "value"})
}

func (suite *ThemeClientTestSuite) TestDeleteAsset() {
	server := suite.NewTestServer(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(suite.T(), "DELETE", r.Method)

		decoder := json.NewDecoder(r.Body)
		var t map[string]Asset
		decoder.Decode(&t)
		defer r.Body.Close()

		assert.Equal(suite.T(), Asset{Key: "deletekey", Value: "value"}, t["asset"])
		fmt.Fprintf(w, jsonFixture("responses/asset"))
	})
	defer server.Close()

	suite.client.DeleteAsset(Asset{Key: "deletekey", Value: "value"})
}

func (suite *ThemeClientTestSuite) TestPerform() {
	server := suite.NewTestServer(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(suite.T(), "POST", r.Method)

		decoder := json.NewDecoder(r.Body)
		var t map[string]Asset
		decoder.Decode(&t)
		defer r.Body.Close()

		assert.Equal(suite.T(), Asset{Key: "fookey", Value: "value"}, t["asset"])
		fmt.Fprintf(w, jsonFixture("responses/asset"))
	})
	defer server.Close()

	suite.client.Perform(Asset{Key: "fookey", Value: "value"}, Create)

	_, err := suite.client.Perform(Asset{Key: "fookeybee", Value: "value"}, Create)
	assert.NotNil(suite.T(), err)
}

func (suite *ThemeClientTestSuite) NewTestServer(handler http.HandlerFunc) *httptest.Server {
	server := httptest.NewServer(handler)
	suite.client.httpClient.config.Domain = server.URL
	return server
}

func TestThemeClientTestSuite(t *testing.T) {
	suite.Run(t, new(ThemeClientTestSuite))
}
