package cmd

import (
	"sync"
	"testing"

	"github.com/spf13/cobra"
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
	environment = "development"

	allenvs = true
	clients, err = generateThemeClients()
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), 3, len(clients))
	allenvs = false

	configPath = badEnvirontmentPath
	_, err = generateThemeClients()
	assert.NotNil(suite.T(), err)
}

func (suite *ThemeTestSuite) TestForEachClient() {
	configPath = goodEnvirontmentPath
	allenvs = true
	runtimes := make(chan int, 100)
	callbacks := make(chan int, 100)
	runner := forEachClient(func(client kit.ThemeClient, filenames []string, wg *sync.WaitGroup) {
		defer wg.Done()
		runtimes <- 1
	}, func(client kit.ThemeClient, filenames []string, wg *sync.WaitGroup) {
		callbacks <- 1
	})
	assert.NotNil(suite.T(), runner)

	err := runner(&cobra.Command{}, []string{})
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), 3, len(runtimes))
	assert.Equal(suite.T(), 3, len(callbacks))

	configPath = badEnvirontmentPath
	err = runner(&cobra.Command{}, []string{})
	assert.NotNil(suite.T(), err)

	allenvs = false
}

func (suite *ThemeTestSuite) TestSetFlagConfig() {
	flagConfig.Password = "foo"
	flagConfig.Domain = "bar"
	flagConfig.Directory = "my/dir/now"
	setFlagConfig()

	config, _ := kit.NewConfiguration()
	assert.Equal(suite.T(), "foo", config.Password)
	assert.Equal(suite.T(), "bar", config.Domain)
	assert.Equal(suite.T(), "my/dir/now", config.Directory)
}

func TestThemeTestSuite(t *testing.T) {
	suite.Run(t, new(ThemeTestSuite))
}
