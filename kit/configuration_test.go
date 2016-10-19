package kit

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type ConfigurationTestSuite struct {
	suite.Suite
}

func (suite *ConfigurationTestSuite) TearDownTest() {
	environmentConfig = Configuration{}
	flagConfig = Configuration{}
}

func (suite *ConfigurationTestSuite) TestSetFlagConfig() {
	config, _ := NewConfiguration()
	assert.Equal(suite.T(), defaultConfig, config)

	flagConfig := Configuration{
		Directory:  "my/dir/now",
		BucketSize: 100,
		RefillRate: 100,
		Timeout:    DefaultTimeout,
	}
	SetFlagConfig(flagConfig)

	config, _ = NewConfiguration()
	assert.Equal(suite.T(), flagConfig, config)
}

func (suite *ConfigurationTestSuite) TestEnvConfig() {
	config, _ := NewConfiguration()
	assert.Equal(suite.T(), defaultConfig, config)

	environmentConfig = Configuration{
		Password:     "password",
		ThemeID:      "themeid",
		Domain:       "nope.myshopify.com",
		Directory:    "my/dir",
		IgnoredFiles: []string{"one", "two", "three"},
		BucketSize:   100,
		RefillRate:   100,
		Proxy:        ":3000",
		Ignores:      []string{"four", "five", "six"},
		Timeout:      40 * time.Second,
	}

	config, _ = NewConfiguration()
	assert.Equal(suite.T(), environmentConfig, config)
}

func (suite *ConfigurationTestSuite) TestConfigPrecedence() {
	config := Configuration{Password: "file"}
	config, _ = config.compile()
	assert.Equal(suite.T(), "file", config.Password)

	environmentConfig = Configuration{Password: "environment"}
	config, _ = config.compile()
	assert.Equal(suite.T(), "environment", config.Password)

	flagConfig = Configuration{Password: "flag"}
	config, _ = config.compile()
	assert.Equal(suite.T(), "flag", config.Password)
}

func (suite *ConfigurationTestSuite) TestValidate() {
	config := Configuration{Password: "file", ThemeID: "123", Domain: "test.myshopify.com"}
	assert.Nil(suite.T(), config.Validate())

	config = Configuration{Password: "file", ThemeID: "live", Domain: "test.myshopify.com"}
	assert.Nil(suite.T(), config.Validate())

	config = Configuration{ThemeID: "123", Domain: "test.myshopify.com"}
	err := config.Validate()
	if assert.NotNil(suite.T(), err) {
		assert.Equal(suite.T(), true, strings.Contains(err.Error(), "missing password"))
	}

	config = Configuration{Password: "test", ThemeID: "123", Domain: "test.nope.com"}
	err = config.Validate()
	if assert.NotNil(suite.T(), err) {
		assert.Equal(suite.T(), true, strings.Contains(err.Error(), "invalid domain"))
	}

	config = Configuration{Password: "test", ThemeID: "123"}
	err = config.Validate()
	if assert.NotNil(suite.T(), err) {
		assert.Equal(suite.T(), true, strings.Contains(err.Error(), "missing domain"))
	}

	config = Configuration{Password: "file", Domain: "test.myshopify.com"}
	err = config.Validate()
	if assert.NotNil(suite.T(), err) {
		assert.Equal(suite.T(), true, strings.Contains(err.Error(), "missing theme_id"))
	}

	config = Configuration{Password: "file", ThemeID: "abc", Domain: "test.myshopify.com"}
	err = config.Validate()
	if assert.NotNil(suite.T(), err) {
		assert.Equal(suite.T(), true, strings.Contains(err.Error(), "invalid theme_id"))
	}
}

func (suite *ConfigurationTestSuite) TestIsLive() {
	config := Configuration{ThemeID: "123"}
	assert.Equal(suite.T(), false, config.IsLive())

	config = Configuration{ThemeID: "live"}
	assert.Equal(suite.T(), true, config.IsLive())
}

func TestConfigurationTestSuite(t *testing.T) {
	suite.Run(t, new(ConfigurationTestSuite))
}
