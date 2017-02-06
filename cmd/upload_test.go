package cmd

import (
	"encoding/json"
	"net/http"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/Shopify/themekit/kit"
)

type UploadTestSuite struct {
	suite.Suite
}

func (suite *UploadTestSuite) SetupTest() {
	configPath = "../fixtures/project/valid_config.yml"
}

func (suite *UploadTestSuite) TestUploadWithFilenames() {
	client, server := newClientAndTestServer(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(suite.T(), "PUT", r.Method)

		var t map[string]kit.Asset
		json.NewDecoder(r.Body).Decode(&t)
		defer r.Body.Close()

		assert.Equal(suite.T(), kit.Asset{Key: "templates/template.liquid", Value: ""}, t["asset"])
	})
	defer server.Close()

	var wg sync.WaitGroup
	wg.Add(1)
	go upload(client, []string{"templates/template.liquid"}, &wg)
	wg.Wait()
}

func (suite *UploadTestSuite) TestUploadWithDirectoryNames() {
	reqCount := make(chan int, 100)
	client, server := newClientAndTestServer(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(suite.T(), "PUT", r.Method)
		reqCount <- 1
	})
	defer server.Close()

	var wg sync.WaitGroup
	wg.Add(1)
	go upload(client, []string{"templates"}, &wg)
	wg.Wait()
	assert.Equal(suite.T(), 2, len(reqCount))
}

func (suite *UploadTestSuite) TestUploadWithBadFileNames() {
	reqCount := make(chan int, 100)
	client, server := newClientAndTestServer(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(suite.T(), "PUT", r.Method)
		reqCount <- 1
	})
	defer server.Close()

	var wg sync.WaitGroup
	wg.Add(1)
	go upload(client, []string{"templates/foo.liquid"}, &wg)
	wg.Wait()
	assert.Equal(suite.T(), 0, len(reqCount))
}

func (suite *UploadTestSuite) TestUploadAll() {
	requests := map[string]string{}
	client, server := newClientAndTestServer(func(w http.ResponseWriter, r *http.Request) {
		var t map[string]kit.Asset
		json.NewDecoder(r.Body).Decode(&t)
		defer r.Body.Close()
		requests[t["asset"].Key] = r.Method
	})
	defer server.Close()

	var wg sync.WaitGroup
	wg.Add(1)
	go upload(client, []string{}, &wg)
	wg.Wait()

	assets, _ := client.LocalAssets()
	assert.Equal(suite.T(), len(assets)-1, len(requests))
	for _, asset := range assets {
		if asset.Key == settingsDataKey {
			continue
		}
		assert.Equal(suite.T(), "PUT", requests[asset.Key])
	}
}

func (suite *UploadTestSuite) TestReadOnlyUpload() {
	requested := false
	client, server := newClientAndTestServer(func(w http.ResponseWriter, r *http.Request) {
		requested = true
	})
	defer server.Close()

	var wg sync.WaitGroup
	wg.Add(1)
	client.Config.ReadOnly = true
	go upload(client, []string{}, &wg)
	wg.Wait()

	assert.Equal(suite.T(), false, requested)
}

func (suite *UploadTestSuite) TestUploadSettingsData() {
	requests := make(chan int, 100)
	client, server := newClientAndTestServer(func(w http.ResponseWriter, r *http.Request) {
		requests <- 1
	})
	defer server.Close()

	var wg sync.WaitGroup
	uploadSettingsData(client, []string{}, &wg)
	uploadSettingsData(client, []string{"templates/template.liquid"}, &wg)
	uploadSettingsData(client, []string{"templates/template.liquid", "config/settings_data.json"}, &wg)
	wg.Wait()
	assert.Equal(suite.T(), 2, len(requests))
}

func TestUploadTestSuite(t *testing.T) {
	suite.Run(t, new(UploadTestSuite))
}
