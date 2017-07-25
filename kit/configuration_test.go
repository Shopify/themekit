package kit

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func resetConfig() {
	flagConfig = Configuration{}
	environmentConfig = Configuration{}
}

func TestSetFlagConfig(t *testing.T) {
	defer resetConfig()

	config, _ := NewConfiguration()
	assert.Equal(t, defaultConfig, *config)

	flagConfig := Configuration{
		Directory: "my/dir/now",
		Timeout:   DefaultTimeout,
	}
	SetFlagConfig(flagConfig)

	config, _ = NewConfiguration()
	assert.Equal(t, flagConfig, *config)

}

func TestConfiguration_Env(t *testing.T) {
	defer resetConfig()

	config, _ := NewConfiguration()
	assert.Equal(t, defaultConfig, *config)

	environmentConfig = Configuration{
		Password:     "password",
		ThemeID:      "themeid",
		Domain:       "nope.myshopify.com",
		Directory:    "my/dir",
		IgnoredFiles: []string{"one", "two", "three"},
		Proxy:        ":3000",
		Ignores:      []string{"four", "five", "six"},
		Timeout:      40 * time.Second,
	}

	config, _ = NewConfiguration()
	assert.Equal(t, environmentConfig, *config)
}

func TestConfiguration_Precedence(t *testing.T) {
	defer resetConfig()

	config := &Configuration{Password: "file"}
	config, _ = config.compile()
	assert.Equal(t, "file", config.Password)

	environmentConfig = Configuration{Password: "environment"}
	config, _ = config.compile()
	assert.Equal(t, "environment", config.Password)

	flagConfig = Configuration{Password: "flag"}
	config, _ = config.compile()
	assert.Equal(t, "flag", config.Password)
}

func TestConfiguration_Validate(t *testing.T) {
	defer resetConfig()

	config := Configuration{Password: "file", ThemeID: "123", Domain: "test.myshopify.com"}
	assert.Nil(t, config.Validate())

	config = Configuration{Password: "file", ThemeID: "live", Domain: "test.myshopify.com"}
	assert.Nil(t, config.Validate())

	config = Configuration{ThemeID: "123", Domain: "test.myshopify.com"}
	err := config.Validate()
	if assert.NotNil(t, err) {
		assert.True(t, strings.Contains(err.Error(), "missing password"))
	}

	config = Configuration{Password: "test", ThemeID: "123", Domain: "test.nope.com"}
	err = config.Validate()
	if assert.NotNil(t, err) {
		assert.True(t, strings.Contains(err.Error(), "invalid store domain"))
	}

	config = Configuration{Password: "test", ThemeID: "123"}
	err = config.Validate()
	if assert.NotNil(t, err) {
		assert.True(t, strings.Contains(err.Error(), "missing store domain"))
	}

	config = Configuration{Password: "file", Domain: "test.myshopify.com"}
	err = config.Validate()
	if assert.NotNil(t, err) {
		assert.True(t, strings.Contains(err.Error(), "missing theme_id"))
	}

	config = Configuration{Password: "file", ThemeID: "abc", Domain: "test.myshopify.com"}
	err = config.Validate()
	if assert.NotNil(t, err) {
		assert.True(t, strings.Contains(err.Error(), "invalid theme_id"))
	}
}

func TestConfiguration_IsLive(t *testing.T) {
	defer resetConfig()

	config := Configuration{ThemeID: "123"}
	assert.False(t, config.IsLive())

	config = Configuration{ThemeID: "live"}
	assert.True(t, config.IsLive())
}

func TestConfiguration_AsYaml(t *testing.T) {
	defer resetConfig()

	config := Configuration{Directory: defaultConfig.Directory}
	assert.Equal(t, "", config.asYAML().Directory)

	config = Configuration{Directory: "nope"}
	assert.Equal(t, "nope", config.asYAML().Directory)

	config = Configuration{Timeout: defaultConfig.Timeout}
	assert.Equal(t, time.Duration(0), config.asYAML().Timeout)

	config = Configuration{Timeout: 42}
	assert.Equal(t, time.Duration(42), config.asYAML().Timeout)
}
