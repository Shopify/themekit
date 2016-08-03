package themekit

import (
	"bytes"
	"net/http"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadingAValidConfiguration(t *testing.T) {
	config, err := LoadConfiguration([]byte(validConfiguration))
	assert.Nil(t, err)
	assert.Equal(t, "example.myshopify.com", config.Domain)
	assert.Equal(t, "abracadabra", config.Password)
	assert.Equal(t, "https://example.myshopify.com/admin", config.URL)
	assert.Equal(t, "https://example.myshopify.com/admin/assets.json", config.AssetPath())
	assert.Equal(t, 4, config.Concurrency)
	assert.Nil(t, config.IgnoredFiles)
}

func TestLoadingAValidConfigurationWithIgnoredFiles(t *testing.T) {
	config, err := LoadConfiguration([]byte(validConfigurationWithIgnoredFiles))
	assert.Nil(t, err)
	assert.Equal(t, "example.myshopify.com", config.Domain)
	assert.Equal(t, "abracadabra", config.Password)
	assert.Equal(t, []string{"charmander", "bulbasaur", "squirtle"}, config.IgnoredFiles)
}

func TestLoadingAValidConfigurationWithAThemeId(t *testing.T) {
	config, err := LoadConfiguration([]byte(validConfigurationWithThemeID))
	assert.Nil(t, err)
	assert.Equal(t, 1234, config.ThemeID)
	assert.Equal(t, "https://example.myshopify.com/admin/themes/1234", config.URL)
	assert.Equal(t, "https://example.myshopify.com/admin/themes/1234/assets.json", config.AssetPath())
}

func TestLoadingSupportedConfiguration(t *testing.T) {
	config, err := LoadConfiguration([]byte(supportedConfiguration))
	assert.Nil(t, err)
	assert.Equal(t, "example.myshopify.com", config.Domain)
	assert.Equal(t, "abracadabra", config.Password)
}

func TestLoadingConfigurationWithMissingFields(t *testing.T) {
	tests := []struct {
		src, expectedError string
	}{
		{configurationWithoutAccessTokenAndPassword, "missing password or access_token (using 'password' is encouraged. 'access_token', which does the same thing will be deprecated soon)"},
		{configurationWithoutDomain, "missing domain"},
		{configurationWithInvalidDomain, "invalid domain, must end in '.myshopify.com'"},
	}

	for _, data := range tests {
		_, err := LoadConfiguration([]byte(data.src))
		assert.NotNil(t, err)
		assert.Equal(t, data.expectedError, err.Error())
	}
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

func TestAddHeadersAddsPlatformAndArchitecture(t *testing.T) {
	req, _ := http.NewRequest("GET", "/foo/bar", nil)

	config := Configuration{Domain: "hello.myshopify.com", AccessToken: "secret", BucketSize: 10, RefillRate: 4}
	config.AddHeaders(req)

	userAgent := req.Header.Get("User-Agent")
	assert.True(t, strings.Contains(userAgent, runtime.GOOS))
	assert.True(t, strings.Contains(userAgent, runtime.GOARCH))
}

const (
	validConfiguration = `
  store: example.myshopify.com
  password: abracadabra
  concurrency: 4
  `

	validConfigurationWithThemeID = `
  store: example.myshopify.com
  password: abracadabra
  theme_id: 1234
  `

	validConfigurationWithIgnoredFiles = `
  store: example.myshopify.com
  password: abracadabra
  ignore_files:
    - charmander
    - bulbasaur
    - squirtle
  `

	supportedConfiguration = `
  store: example.myshopify.com
  password: abracadabra
  theme_id: 12345
  `

	configurationWithoutAccessTokenAndPassword = `
  store: foo.myshopify.com
  theme_id: 123
  `

	configurationWithoutDomain = `
  password: foobar
  `

	configurationWithInvalidDomain = `
  store: example.something.net
  password: abracadabra
  theme_id: 12345
  `
)
