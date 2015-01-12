package phoenix

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestLoadingAValidConfiguration(t *testing.T) {
	config, err := LoadConfiguration([]byte(validConfiguration))
	assert.Nil(t, err)
	assert.Equal(t, "example.myshopify.com", config.Domain)
	assert.Equal(t, "abracadabra", config.AccessToken)
	assert.Equal(t, "https://example.myshopify.com/admin", config.Url)
	assert.Equal(t, "https://example.myshopify.com/admin/assets.json", config.AssetPath())
	assert.Equal(t, 4, config.Concurrency)
	assert.Nil(t, config.IgnoredFiles)
}

func TestLoadingAValidConfigurationWithIgnoredFiles(t *testing.T) {
	config, err := LoadConfiguration([]byte(validConfigurationWithIgnoredFiles))
	assert.Nil(t, err)
	assert.Equal(t, "example.myshopify.com", config.Domain)
	assert.Equal(t, "abracadabra", config.AccessToken)
	assert.Equal(t, []string{"charmander", "bulbasaur", "squirtle"}, config.IgnoredFiles)
}

func TestLoadingAValidConfigurationWithAThemeId(t *testing.T) {
	config, err := LoadConfiguration([]byte(validConfigurationWithThemeId))
	assert.Nil(t, err)
	assert.Equal(t, 1234, config.ThemeId)
	assert.Equal(t, "https://example.myshopify.com/admin/themes/1234", config.Url)
	assert.Equal(t, "https://example.myshopify.com/admin/themes/1234/assets.json", config.AssetPath())
}

func TestLoadingAnUnsupportedConfiguration(t *testing.T) {
	config, err := LoadConfiguration([]byte(unsupportedConfiguration))
	assert.Nil(t, err)
	assert.Equal(t, "example.myshopify.com", config.Domain)
	assert.Equal(t, "abracadabra", config.AccessToken)
}

func TestWritingAConfigurationFile(t *testing.T) {
	buffer := new(bytes.Buffer)
	config := Configuration{Domain: "hello.myshopify.com", AccessToken: "secret", BucketSize: 10, RefillRate: 4}
	err := config.Write(buffer)
	expectedConfiguration :=
		`access_token: secret
store: hello.myshopify.com
bucket_size: 10
refill_rate: 4
`
	assert.Nil(t, err, "An error should not have been raised")
	assert.Equal(t, expectedConfiguration, string(buffer.Bytes()))
}

const (
	validConfiguration = `
  store: example.myshopify.com
  access_token: abracadabra
  concurrency: 4
  `

	validConfigurationWithThemeId = `
  store: example.myshopify.com
  access_token: abracadabra
  theme_id: 1234
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
