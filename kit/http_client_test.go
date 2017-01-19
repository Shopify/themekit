package kit

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type HTTPClientTestSuite struct {
	suite.Suite
	config *Configuration
	client *httpClient
}

func (suite *HTTPClientTestSuite) SetupTest() {
	suite.config, _ = NewConfiguration()
	suite.config.Domain = "test.myshopify.com"
	suite.config.ThemeID = "123"
	suite.config.Password = "sharknado"
	suite.client, _ = newHTTPClient(suite.config)
}

func (suite *HTTPClientTestSuite) TestNewHttpClient() {
	assert.Equal(suite.T(), suite.config, suite.client.config)
	assert.Equal(suite.T(), suite.config.Timeout, suite.client.client.Timeout)

	config, _ := NewConfiguration()
	config.Proxy = "://abc!21@"
	client, err := newHTTPClient(config)
	assert.NotNil(suite.T(), err)

	config, _ = NewConfiguration()
	config.Proxy = "http://localhost:3000"
	client, err = newHTTPClient(config)
	assert.Nil(suite.T(), err)
	assert.NotNil(suite.T(), client.client.Transport)
}

func (suite *HTTPClientTestSuite) TestAdminURL() {
	assert.Equal(suite.T(),
		fmt.Sprintf("https://%s/admin/themes/%v", suite.config.Domain, suite.config.ThemeID),
		suite.client.AdminURL())

	suite.client.config.ThemeID = "live"

	assert.Equal(suite.T(),
		fmt.Sprintf("https://%s/admin", suite.config.Domain),
		suite.client.AdminURL())
}

func (suite *HTTPClientTestSuite) TestAssetPath() {
	assert.Equal(suite.T(),
		fmt.Sprintf("%s/assets.json", suite.client.AdminURL()),
		suite.client.AssetPath())
}

func (suite *HTTPClientTestSuite) TestThemesPath() {
	assert.Equal(suite.T(),
		fmt.Sprintf("%s/themes.json", suite.client.AdminURL()),
		suite.client.ThemesPath())
}

func (suite *HTTPClientTestSuite) TestThemePath() {
	assert.Equal(suite.T(),
		fmt.Sprintf("%s/themes/456.json", suite.client.AdminURL()),
		suite.client.ThemePath(456))
}

func (suite *HTTPClientTestSuite) TestAssetQuery() {
	server := suite.NewTestServer(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(suite.T(), "GET", r.Method)
		assert.Equal(suite.T(), "", r.URL.RawQuery)
		fmt.Fprintf(w, jsonFixture("responses/assets"))
	})
	resp, err := suite.client.AssetQuery(Retrieve, map[string]string{})
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), listRequest, resp.Type)
	assert.Equal(suite.T(), 2, len(resp.Assets))
	server.Close()

	server = suite.NewTestServer(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(suite.T(), "GET", r.Method)
		assert.Equal(suite.T(), "asset[key]=file.txt", r.URL.RawQuery)
		fmt.Fprintf(w, jsonFixture("responses/asset"))
	})
	resp, err = suite.client.AssetQuery(Retrieve, map[string]string{"asset[key]": "file.txt"})
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), assetRequest, resp.Type)
	assert.Equal(suite.T(), "assets/hello.txt", resp.Asset.Key)
	server.Close()
}

func (suite *HTTPClientTestSuite) TestNewTheme() {
	server := suite.NewTestServer(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(suite.T(), "POST", r.Method)

		decoder := json.NewDecoder(r.Body)
		var t map[string]Theme
		decoder.Decode(&t)
		defer r.Body.Close()

		assert.Equal(suite.T(), Theme{Name: "name", Source: "source", Role: "unpublished"}, t["theme"])
		fmt.Fprintf(w, jsonFixture("responses/theme"))
	})
	defer server.Close()
	resp, err := suite.client.NewTheme("name", "source")
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), themeRequest, resp.Type)
	assert.Equal(suite.T(), "timberland", resp.Theme.Name)
}

func (suite *HTTPClientTestSuite) TestGetTheme() {
	server := suite.NewTestServer(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(suite.T(), "GET", r.Method)
		fmt.Fprintf(w, jsonFixture("responses/theme"))
	})
	resp, err := suite.client.GetTheme(123)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), themeRequest, resp.Type)
	assert.Equal(suite.T(), "timberland", resp.Theme.Name)
	server.Close()
}

func (suite *HTTPClientTestSuite) TestAssetAction() {
	server := suite.NewTestServer(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(suite.T(), "PUT", r.Method)

		decoder := json.NewDecoder(r.Body)
		var t map[string]Asset
		decoder.Decode(&t)
		defer r.Body.Close()

		assert.Equal(suite.T(), Asset{Key: "key", Value: "value"}, t["asset"])
		fmt.Fprintf(w, jsonFixture("responses/asset"))
	})
	defer server.Close()
	resp, err := suite.client.AssetAction(Update, Asset{Key: "key", Value: "value"})
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), assetRequest, resp.Type)
	assert.Equal(suite.T(), "key", resp.Asset.Key)
}

func (suite *HTTPClientTestSuite) TestNewRequest() {
	req, err := suite.client.newRequest(Update, suite.config.Domain, nil)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), suite.config.Password, req.Header.Get("X-Shopify-Access-Token"))
	assert.Equal(suite.T(), "application/json", req.Header.Get("Content-Type"))
	assert.Equal(suite.T(), "application/json", req.Header.Get("Accept"))
	assert.Equal(suite.T(), fmt.Sprintf("go/themekit (%s; %s)", runtime.GOOS, runtime.GOARCH), req.Header.Get("User-Agent"))

	_, err = suite.client.newRequest(Update, "://#nksd", nil)
	assert.NotNil(suite.T(), err)
}

func (suite *HTTPClientTestSuite) TestSendJSON() {
	server := suite.NewTestServer(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(suite.T(), "PUT", r.Method)

		decoder := json.NewDecoder(r.Body)
		var t map[string]string
		decoder.Decode(&t)
		defer r.Body.Close()

		assert.Equal(suite.T(), "mystring", t["test"])
	})
	defer server.Close()
	resp, err := suite.client.sendJSON(assetRequest, Update, server.URL, map[string]interface{}{"test": "mystring"})
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), assetRequest, resp.Type)
}

func (suite *HTTPClientTestSuite) TestSendRequest() {
	server := suite.NewTestServer(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(suite.T(), "PUT", r.Method)

		buf := new(bytes.Buffer)
		buf.ReadFrom(r.Body)

		assert.Equal(suite.T(), "my string", buf.String())
	})
	defer server.Close()
	resp, err := suite.client.sendRequest(assetRequest, Update, server.URL, bytes.NewBufferString("my string"))
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), assetRequest, resp.Type)
}

func (suite *HTTPClientTestSuite) NewTestServer(handler http.HandlerFunc) *httptest.Server {
	server := httptest.NewServer(handler)
	suite.client.config.Domain = server.URL
	return server
}

func TestHttpClientTestSuite(t *testing.T) {
	suite.Run(t, new(HTTPClientTestSuite))
}

func fileFixture(name string) *os.File {
	path := fmt.Sprintf("../fixtures/%s.json", name)
	file, _ := os.Open(path)
	return file
}

func jsonFixture(name string) string {
	bytes, err := ioutil.ReadAll(fileFixture(name))
	if err != nil {
		log.Fatal(err)
	}
	return string(bytes)
}
