package cmd

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Shopify/themekit/kit"
	"github.com/Shopify/themekit/kittest"
)

func TestDeploy(t *testing.T) {
	server := kittest.NewTestServer()
	defer server.Close()
	assert.Nil(t, kittest.GenerateConfig(server.URL, true))
	kittest.TouchFixtureFile(filepath.Join("config", "settings_data.json"), "")
	defer kittest.Cleanup()
	defer resetArbiter()

	client, err := getClient()
	if assert.Nil(t, err) {
		deployMthd := deploy(true)
		client.Config.ReadOnly = true
		err := deployMthd(client, []string{})
		assert.True(t, strings.Contains(err.Error(), "environment is readonly"))
		client.Config.ReadOnly = false

		err = deployMthd(client, []string{})
		assert.True(t, strings.Contains(err.Error(), "Diff"))

		arbiter.force = true
		assert.Nil(t, deployMthd(client, []string{}))

		err = deployMthd(client, []string{"nope.txt"})
		assert.True(t, strings.Contains(err.Error(), "no such file"))
	}
}

func TestIncBar(t *testing.T) {
	incBar(arbiter.progress.AddBar(int64(1)))
}

func TestPerform(t *testing.T) {
	server := kittest.NewTestServer()
	defer server.Close()
	assert.Nil(t, kittest.GenerateConfig(server.URL, true))
	kittest.TouchFixtureFile(filepath.Join("config", "settings_data.json"), "")
	defer kittest.Cleanup()
	defer resetArbiter()

	client, err := getClient()
	if assert.Nil(t, err) {
		server.Reset()

		err := perform(client, kit.Asset{}, kit.Update, nil)
		assert.True(t, strings.Contains(err.Error(), "No collection name provided"))
		err = perform(client, kit.Asset{}, kit.Remove, nil)
		assert.True(t, strings.Contains(err.Error(), "No collection name provided"))

		assert.Nil(t, perform(client, kit.Asset{Key: "asset.js"}, kit.Update, nil))
		arbiter.force = true
		assert.Nil(t, perform(client, kit.Asset{Key: "asset.js"}, kit.Update, nil))
		assert.Nil(t, perform(client, kit.Asset{Key: "asset.js"}, kit.Remove, nil))

		client.Config.Environment = ""
		err = perform(client, kit.Asset{Key: "empty"}, kit.Remove, nil)
		assert.True(t, strings.Contains(err.Error(), "No key name provided"))
	}
}

func TestUploadSettingsData(t *testing.T) {
	server := kittest.NewTestServer()
	defer server.Close()
	assert.Nil(t, kittest.GenerateConfig(server.URL, true))
	kittest.TouchFixtureFile(filepath.Join("config", "settings_data.json"), "")
	defer kittest.Cleanup()
	defer resetArbiter()

	client, err := getClient()
	if assert.Nil(t, err) {
		server.Reset()

		assert.Nil(t, uploadSettingsData(client, []string{}))
		assert.Nil(t, uploadSettingsData(client, []string{"templates/template.liquid"}))
		assert.Nil(t, uploadSettingsData(client, []string{"templates/template.liquid", "config/settings_data.json"}))

		kittest.Cleanup()
		assert.NotNil(t, uploadSettingsData(client, []string{}))

		assert.Equal(t, 2, len(server.Requests))
	}
}

func TestIndexOf(t *testing.T) {
	array := []string{"one", "two", "three"}

	assert.Equal(t, 0, indexOf(len(array), func(i int) bool { return array[i] == "one" }))
	assert.Equal(t, 1, indexOf(len(array), func(i int) bool { return array[i] == "two" }))
	assert.Equal(t, 2, indexOf(len(array), func(i int) bool { return array[i] == "three" }))
	assert.Equal(t, -1, indexOf(len(array), func(i int) bool { return array[i] == "four" }))
}
