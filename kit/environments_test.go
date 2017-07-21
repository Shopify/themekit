package kit

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Shopify/themekit/kittest"
)

func TestLoadEnvironments(t *testing.T) {
	defer kittest.Cleanup()

	kittest.GenerateConfig("example.myshopify.io", true)
	envs, err := LoadEnvironments("config.yml")
	assert.Nil(t, err)
	assert.Equal(t, 3, len(envs))

	kittest.GenerateConfig("example.myshopify.io", false)
	_, err = LoadEnvironments("config.yml")
	assert.NotNil(t, err)

	_, err = LoadEnvironments("nope.yml")
	assert.NotNil(t, err)

	kittest.TouchFixtureFile("config.json", ":this is not json")
	_, err = LoadEnvironments(filepath.Join(kittest.FixtureProjectPath, "config.json"))
	assert.NotNil(t, err)
}

func TestSearchConfigPath(t *testing.T) {
	defer kittest.Cleanup()
	kittest.GenerateConfig("example.myshopify.io", true)

	_, ext, err := searchConfigPath("config.yml")
	assert.Nil(t, err)
	assert.Equal(t, "yml", ext)

	kittest.Cleanup()
	kittest.GenerateJSONConfig("example.myshopify.io")
	_, ext, err = searchConfigPath("config.json")
	assert.Nil(t, err)
	assert.Equal(t, "json", ext)

	_, _, err = searchConfigPath("not_there.yml")
	assert.NotNil(t, err)
	assert.Equal(t, os.ErrNotExist, err)
}

func TestSetConfiguration(t *testing.T) {
	defer kittest.Cleanup()
	kittest.GenerateConfig("example.myshopify.io", true)
	envs, err := LoadEnvironments("config.yml")
	assert.Nil(t, err)
	newConfig, _ := NewConfiguration()
	envs.SetConfiguration("test", newConfig)
	assert.Equal(t, newConfig, envs["test"])
}

func TestGetConfiguration(t *testing.T) {
	defer kittest.Cleanup()
	kittest.GenerateConfig("example.myshopify.io", true)
	envs, err := LoadEnvironments("config.yml")
	assert.Nil(t, err)
	_, err = envs.GetConfiguration("development")
	assert.Nil(t, err)
	_, err = envs.GetConfiguration("nope")
	assert.NotNil(t, err)
	envs["test"] = nil
	_, err = envs.GetConfiguration("test")
	assert.NotNil(t, err)
}

func TestSave(t *testing.T) {
	defer kittest.Cleanup()
	kittest.GenerateConfig("example.myshopify.io", true)
	envs, err := LoadEnvironments("config.yml")
	assert.Nil(t, err)
	assert.Nil(t, envs.Save("config.json"))
	_, err = os.Stat("config.json")
	assert.Nil(t, err)
	assert.NotNil(t, envs.Save("./no/where/path"))
}
