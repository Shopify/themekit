package cmd

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Shopify/themekit/kit"
)

func TestSaveConfiguration(t *testing.T) {
	configPath = goodEnvirontmentPath
	env, err := kit.LoadEnvironments(configPath)
	config, _ := env.GetConfiguration(kit.DefaultEnvironment)

	err = saveConfiguration(config)
	assert.Nil(t, err)

	configPath = badEnvirontmentPath
	err = saveConfiguration(config)
	assert.NotNil(t, err)

	configPath = "../fixtures/project/out.yml"
	err = saveConfiguration(config)
	assert.Nil(t, err)

	os.Remove(configPath)
}
