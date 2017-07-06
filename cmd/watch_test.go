package cmd

// import (
//	"encoding/json"
//	"net/http"
//	"os"
//	"testing"

//	"github.com/stretchr/testify/assert"
//	"github.com/stretchr/testify/suite"

//	"github.com/Shopify/themekit/kit"
// )

// type WatchTestSuite struct {
//	suite.Suite
// }

// func (suite *WatchTestSuite) TestStartWatch() {
//	go func() {
//		signalChan <- os.Interrupt
//	}()
//	err := startWatch()
//	assert.NotNil(suite.T(), err)
// }

// func (suite *WatchTestSuite) TestWatch() {
//	client, server := newClientAndTestServer(func(w http.ResponseWriter, r *http.Request) {})
//	defer server.Close()
//	go func() {
//		signalChan <- os.Interrupt
//	}()
//	watch([]kit.ThemeClient{client})
// }

// func (suite *WatchTestSuite) TestReadOnlyWatch() {
//	requested := false
//	client, server := newClientAndTestServer(func(w http.ResponseWriter, r *http.Request) {
//		requested = true
//	})
//	defer server.Close()
//	client.Config.ReadOnly = true
//	watch([]kit.ThemeClient{client})
//	assert.Equal(suite.T(), false, requested)
// }

// func (suite *WatchTestSuite) TestHandleWatchReload() {
//	client, server := newClientAndTestServer(func(w http.ResponseWriter, r *http.Request) {})
//	defer server.Close()
//	go func() {
//		reloadSignal <- true
//	}()
//	err := watch([]kit.ThemeClient{client})
//	assert.Equal(suite.T(), err, errReload)
// }

// func (suite *WatchTestSuite) TestHandleWatchEvent() {
//	requests := make(chan int, 1000)
//	client, server := newClientAndTestServer(func(w http.ResponseWriter, r *http.Request) {
//		assert.Equal(suite.T(), "DELETE", r.Method)

//		decoder := json.NewDecoder(r.Body)
//		var t map[string]kit.Asset
//		decoder.Decode(&t)
//		defer r.Body.Close()

//		assert.Equal(suite.T(), kit.Asset{Key: "templates/layout.liquid", Value: ""}, t["asset"])
//		requests <- 1
//	})
//	defer server.Close()

//	handleWatchEvent(client, kit.Asset{Key: "templates/layout.liquid"}, kit.Remove)

//	assert.Equal(suite.T(), 1, len(requests))
// }

// func TestWatchTestSuite(t *testing.T) {
//	suite.Run(t, new(WatchTestSuite))
// }
