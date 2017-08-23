package cmd

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Shopify/themekit/kit"
	"github.com/Shopify/themekit/kittest"
)

func TestStartWatch(t *testing.T) {
	server := kittest.NewTestServer()
	defer server.Close()

	assert.NotNil(t, startWatch(nil, []string{}))

	assert.Nil(t, kittest.GenerateConfig(server.URL, true))
	defer kittest.Cleanup()

	go func() {
		reloadSignal <- true
		signalChan <- os.Interrupt
	}()
	assert.Nil(t, startWatch(nil, []string{}))
}

func TestWatch(t *testing.T) {
	server := kittest.NewTestServer()
	defer server.Close()
	assert.Nil(t, kittest.GenerateConfig(server.URL, true))
	defer kittest.Cleanup()
	defer resetArbiter()

	_, err := getClient()
	if assert.Nil(t, err) {
		go func() { reloadSignal <- true }()
		err := watch()
		assert.Equal(t, errReload, err)

		for _, client := range arbiter.activeThemeClients {
			client.Config.ReadOnly = true
		}
		err = watch()
		assert.Equal(t, err.Error(), "no valid configuration to start watch")
		for _, client := range arbiter.activeThemeClients {
			client.Config.ReadOnly = false
		}

		os.Remove("config.yml")
		err = watch()
		assert.True(t, strings.Contains(err.Error(), "no such file or directory"))

		kittest.Cleanup()
		err = watch()
		assert.Equal(t, err.Error(), "lstat fixtures: no such file or directory")
	}
}

func TestHandleWatchEvent(t *testing.T) {
	server := kittest.NewTestServer()
	defer server.Close()
	assert.Nil(t, kittest.GenerateConfig(server.URL, true))
	defer kittest.Cleanup()
	defer resetArbiter()

	client, err := getClient()
	if assert.Nil(t, err) {
		server.Reset()
		handleWatchEvent(client, kit.Asset{Key: "templates/layout.liquid"}, kit.Remove, nil)
		assert.Equal(t, 1, len(server.Requests))
		assert.Equal(t, "DELETE", server.Requests[0].Method)
		assert.True(t, strings.Contains(stdOutOutput.String(), "Received"))
		assert.True(t, strings.Contains(stdOutOutput.String(), "Successfully"))

		server.Reset()
		resetLog()
		handleWatchEvent(client, kit.Asset{Key: "nope"}, kit.Update, nil)
		assert.Equal(t, 1, len(server.Requests))
		assert.Equal(t, "PUT", server.Requests[0].Method)

		assert.True(t, strings.Contains(stdOutOutput.String(), "Received"))
		assert.True(t, strings.Contains(stdOutOutput.String(), "Conflict"))
	}
}
