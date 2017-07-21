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

func TestEnvConfig(t *testing.T) {
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

func TestConfigPrecedence(t *testing.T) {
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

func TestValidate(t *testing.T) {
	defer resetConfig()

	config := Configuration{Password: "file", ThemeID: "123", Domain: "test.myshopify.com"}
	assert.Nil(t, config.Validate())

	config = Configuration{Password: "file", ThemeID: "live", Domain: "test.myshopify.com"}
	assert.Nil(t, config.Validate())

	config = Configuration{ThemeID: "123", Domain: "test.myshopify.com"}
	err := config.Validate()
	if assert.NotNil(t, err) {
		assert.Equal(t, true, strings.Contains(err.Error(), "missing password"))
	}

	config = Configuration{Password: "test", ThemeID: "123", Domain: "test.nope.com"}
	err = config.Validate()
	if assert.NotNil(t, err) {
		assert.Equal(t, true, strings.Contains(err.Error(), "invalid store domain"))
	}

	config = Configuration{Password: "test", ThemeID: "123"}
	err = config.Validate()
	if assert.NotNil(t, err) {
		assert.Equal(t, true, strings.Contains(err.Error(), "missing store domain"))
	}

	config = Configuration{Password: "file", Domain: "test.myshopify.com"}
	err = config.Validate()
	if assert.NotNil(t, err) {
		assert.Equal(t, true, strings.Contains(err.Error(), "missing theme_id"))
	}

	config = Configuration{Password: "file", ThemeID: "abc", Domain: "test.myshopify.com"}
	err = config.Validate()
	if assert.NotNil(t, err) {
		assert.Equal(t, true, strings.Contains(err.Error(), "invalid theme_id"))
	}
}

func TestIsLive(t *testing.T) {
	defer resetConfig()

	config := Configuration{ThemeID: "123"}
	assert.Equal(t, false, config.IsLive())

	config = Configuration{ThemeID: "live"}
	assert.Equal(t, true, config.IsLive())
}
