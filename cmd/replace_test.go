package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/Shopify/themekit/kit"
)

type ReplaceTestSuite struct {
	suite.Suite
}

func (suite *ReplaceTestSuite) SetupTest() {
	configPath = "../fixtures/project/valid_config.yml"
}

func (suite *ReplaceTestSuite) TestReplaceWithFilenames() {
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
	go replace(client, []string{"templates/template.liquid"}, &wg)
	wg.Wait()
}

func (suite *ReplaceTestSuite) TestReplaceAll() {
	var firstKey string
	assetListServed := false
	requests := map[string]string{}
	client, server := newClientAndTestServer(func(w http.ResponseWriter, r *http.Request) {
		if assetListServed {
			var t map[string]kit.Asset
			json.NewDecoder(r.Body).Decode(&t)
			defer r.Body.Close()
			requests[t["asset"].Key] = r.Method
		} else {
			bytes, _ := json.Marshal(map[string][]kit.Asset{"assets": {{Key: "templates/nope.liquid"}, {Key: firstKey}}})
			fmt.Fprintf(w, string(bytes))
			assetListServed = true
		}
	})
	defer server.Close()

	assets, _ := client.LocalAssets()
	for _, asset := range assets {
		firstKey = asset.Key
		break
	}

	var wg sync.WaitGroup
	wg.Add(1)
	go replace(client, []string{}, &wg)
	wg.Wait()

	assert.Equal(suite.T(), "DELETE", requests["templates/nope.liquid"])
	for _, asset := range assets {
		assert.Equal(suite.T(), "PUT", requests[asset.Key])
	}
}

func TestReplaceTestSuite(t *testing.T) {
	suite.Run(t, new(ReplaceTestSuite))
}
