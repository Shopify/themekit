package cmd

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Shopify/themekit/kittest"
)

func TestRemove(t *testing.T) {
	server := kittest.NewTestServer()
	defer server.Close()
	assert.Nil(t, kittest.GenerateConfig(server.URL, true))
	defer kittest.Cleanup()
	defer resetArbiter()

	client, err := getClient()
	server.Reset()
	if assert.Nil(t, err) {
		arbiter.force = true
		err = remove(client, []string{"templates/layout.liquid"})
		assert.True(t, os.IsNotExist(err))
		if assert.Equal(t, 1, len(server.Requests)) {
			assert.Equal(t, "DELETE", server.Requests[0].Method)
		}

		arbiter.force = false
		err = remove(client, []string{"templates/layout.liquid"})
		assert.True(t, strings.Contains(err.Error(), "file was modified remotely"))

		client.Config.ReadOnly = true
		err := remove(client, []string{"templates/layout.liquid"})
		assert.True(t, strings.Contains(err.Error(), "environment is reaonly"))

		client.Config.ReadOnly = false
		err = remove(client, []string{})
		assert.True(t, strings.Contains(err.Error(), "please specify file(s) to be removed"))

		arbiter.force = true
		server.Close()
		assert.Nil(t, remove(client, []string{"templates/layout.liquid"}))
	}
}
