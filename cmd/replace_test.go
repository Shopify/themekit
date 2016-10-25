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

type ReplaceTestSuite struct {
	suite.Suite
}

func (suite *ReplaceTestSuite) SetupTest() {
	configPath = "../fixtures/project/valid_config.yml"
}

func (suite *ReplaceTestSuite) TestReplace() {
	client, server := newClientAndTestServer(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(suite.T(), "PUT", r.Method)

		decoder := json.NewDecoder(r.Body)
		var t map[string]kit.Asset
		decoder.Decode(&t)
		defer r.Body.Close()

		assert.Equal(suite.T(), kit.Asset{Key: "templates/template.liquid", Value: ""}, t["asset"])
	})
	defer server.Close()

	var wg sync.WaitGroup
	wg.Add(1)
	go replace(client, []string{"templates/template.liquid"}, &wg)
	wg.Wait()
}

func TestReplaceTestSuite(t *testing.T) {
	suite.Run(t, new(ReplaceTestSuite))
}
