package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/Shopify/themekit/kit"
)

const (
	goodEnvirontmentPath = "../fixtures/project/valid_config.yml"
	badEnvirontmentPath  = "../fixtures/project/invalid_config.yml"
)

type ThemeTestSuite struct {
	suite.Suite
}

func (suite *ThemeTestSuite) TestGenerateThemeClients() {
	configPath = goodEnvirontmentPath
	clients, err := generateThemeClients()
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), 1, len(clients))

	environment = "nope"
	clients, err = generateThemeClients()
	assert.NotNil(suite.T(), err)

	allenvs = true
	clients, err = generateThemeClients()
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), 3, len(clients))

	configPath = badEnvirontmentPath
	clients, err = generateThemeClients()
	assert.NotNil(suite.T(), err)
}

func (suite *ThemeTestSuite) TestSetFlagConfig() {
	password = "foo"
	domain = "bar"
	directory = "my/dir/now"
	setFlagConfig()

	config, _ := kit.NewConfiguration()
	assert.Equal(suite.T(), "foo", config.Password)
	assert.Equal(suite.T(), "bar", config.Domain)
	assert.Equal(suite.T(), "my/dir/now", config.Directory)
}

func TestThemeTestSuite(t *testing.T) {
	suite.Run(t, new(ThemeTestSuite))
}
