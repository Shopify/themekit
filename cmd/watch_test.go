package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/Shopify/themekit/kit"
)

type WatchTestSuite struct {
	suite.Suite
}

func (suite *WatchTestSuite) TestReloadableWatch() {
}

func (suite *WatchTestSuite) TestWatch() {
	client, server := newClientAndTestServer(func(w http.ResponseWriter, r *http.Request) {})
	defer server.Close()
	go func() {
		signalChan <- os.Interrupt
	}()
	watch([]kit.ThemeClient{client})
}

func (suite *WatchTestSuite) TestHandleWatchEvent() {
	requests := make(chan int, 1000)
	client, server := newClientAndTestServer(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(suite.T(), "DELETE", r.Method)

		decoder := json.NewDecoder(r.Body)
		var t map[string]kit.Asset
		decoder.Decode(&t)
		defer r.Body.Close()

		assert.Equal(suite.T(), kit.Asset{Key: "templates/layout.liquid", Value: ""}, t["asset"])
		requests <- 1
	})
	defer server.Close()

	handleWatchEvent(client, kit.Asset{Key: "templates/layout.liquid"}, kit.Remove, fmt.Errorf("bad watch event"))

	handleWatchEvent(client, kit.Asset{Key: "templates/layout.liquid"}, kit.Remove, nil)

	configPath = "/this/is/config/path.yml"
	signalChan = make(chan os.Signal, 100)

	handleWatchEvent(client, kit.Asset{Key: configPath}, kit.Update, fmt.Errorf("not in project"))
	assert.Equal(suite.T(), true, isReloading)
	assert.Equal(suite.T(), 1, len(signalChan))

	configPath = ""
	signalChan = make(chan os.Signal)

	assert.Equal(suite.T(), 1, len(requests))
}

func TestWatchTestSuite(t *testing.T) {
	suite.Run(t, new(WatchTestSuite))
}
