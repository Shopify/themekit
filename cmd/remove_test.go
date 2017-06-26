package cmd

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/Shopify/themekit/kit"
)

type RemoveTestSuite struct {
	suite.Suite
}

func (suite *RemoveTestSuite) TestRemove() {
	client, server := newClientAndTestServer(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(suite.T(), "DELETE", r.Method)

		decoder := json.NewDecoder(r.Body)
		var t map[string]kit.Asset
		decoder.Decode(&t)
		defer r.Body.Close()

		assert.Equal(suite.T(), kit.Asset{Key: "templates/layout.liquid", Value: ""}, t["asset"])
	})
	defer server.Close()

	var wg sync.WaitGroup
	wg.Add(1)
	go remove(client, []string{"templates/layout.liquid"}, &wg)
	wg.Wait()
}

func (suite *RemoveTestSuite) TestReadOnlyRemove() {
	requested := false
	client, server := newClientAndTestServer(func(w http.ResponseWriter, r *http.Request) {
		requested = true
	})
	defer server.Close()

	var wg sync.WaitGroup
	wg.Add(1)
	client.Config.ReadOnly = true
	go remove(client, []string{"templates/layout.liquid"}, &wg)
	wg.Wait()

	assert.Equal(suite.T(), false, requested)
}

func TestRemoveTestSuite(t *testing.T) {
	suite.Run(t, new(RemoveTestSuite))
}

func newTestClient(domain string) kit.ThemeClient {
	config, _ := kit.NewConfiguration()
	config.Environment = "test"
	config.Domain = domain
	config.ThemeID = "123"
	config.Password = "sharknado"
	config.Directory = "../fixtures/project"
	client, _ := kit.NewThemeClient(config)
	return client
}

func newClientAndTestServer(handler http.HandlerFunc) (kit.ThemeClient, *httptest.Server) {
	server := httptest.NewServer(handler)
	return newTestClient(server.URL), server
}
