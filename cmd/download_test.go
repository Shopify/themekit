package cmd

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Shopify/themekit/kittest"
)

func TestDownload(t *testing.T) {
	server := kittest.NewTestServer()
	defer server.Close()
	assert.Nil(t, kittest.GenerateConfig(server.URL, true))
	defer kittest.Cleanup()

	client, err := getClient()
	if assert.Nil(t, err) {
		assert.Nil(t, download(client, []string{"assets/hello.txt"}))
	}
}

func TestDownloadWithFile(t *testing.T) {
	server := kittest.NewTestServer()
	defer server.Close()
	assert.Nil(t, kittest.GenerateConfig(server.URL, true))
	defer kittest.Cleanup()

	client, err := getClient()
	if assert.Nil(t, err) {
		err := downloadFile(client, "assets/hello.txt", nil, nil)
		assert.NotNil(t, err)
		assert.Equal(t, "Skipping because versions match", err.Error())

		arbiter.force = true
		assert.Nil(t, downloadFile(client, "assets/hello.txt", nil, nil))
		println(client.Config.Directory)

		err = downloadFile(client, "nope.txt", nil, nil)
		assert.NotNil(t, err)
		assert.True(t, strings.Contains(err.Error(), "error downloading asset:"))

		oldDir := client.Config.Directory
		client.Config.Directory = "nonexistant"
		err = downloadFile(client, "assets/hello.txt", nil, nil)
		assert.NotNil(t, err)
		assert.True(t, strings.Contains(err.Error(), "error writing asset: "))

		client.Config.Directory = oldDir
		client.Config.Environment = ""
		err = downloadFile(client, "assets/hello.txt", nil, nil)
		assert.NotNil(t, err)
		assert.True(t, strings.Contains(err.Error(), "error updating manifest:"))
	}
}
