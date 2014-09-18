package phoenix

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestLoadingAValidConfiguration(t *testing.T) {
	config, err := LoadConfiguration([]byte(validConfiguration))
	assert.Nil(t, err)
	assert.Equal(t, "example.myshopify.com", config.Domain)
	assert.Equal(t, "abracadabra", config.AccessToken)
	assert.Equal(t, "https://example.myshopify.com/admin", config.Url)
	assert.Nil(t, config.IgnoredFiles)
}

func TestLoadingAValidConfigurationWithIgnoredFiles(t *testing.T) {
	config, err := LoadConfiguration([]byte(validConfigurationWithIgnoredFiles))
	assert.Nil(t, err)
	assert.Equal(t, "example.myshopify.com", config.Domain)
	assert.Equal(t, "abracadabra", config.AccessToken)
	assert.Equal(t, []string{"charmander", "bulbasaur", "squirtle"}, config.IgnoredFiles)
}

func TestLoadingAnUnsupportedConfiguration(t *testing.T) {
	config, err := LoadConfiguration([]byte(unsupportedConfiguration))
	assert.Nil(t, err)
	assert.Equal(t, "example.myshopify.com", config.Domain)
	assert.Equal(t, "abracadabra", config.AccessToken)
}

const (
	validConfiguration = `
  store: example.myshopify.com
  access_token: abracadabra
  `

	validConfigurationWithIgnoredFiles = `
  store: example.myshopify.com
  access_token: abracadabra
  ignore_files:
    - charmander
    - bulbasaur
    - squirtle
  `

	unsupportedConfiguration = `
  store: example.myshopify.com
  access_token: abracadabra
  theme_id: 12345
  `
)
