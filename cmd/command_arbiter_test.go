package cmd

// import (
//	"sync"
//	"testing"

//	"github.com/spf13/cobra"
//	"github.com/stretchr/testify/assert"
//	"github.com/stretchr/testify/suite"

//	"github.com/Shopify/themekit/kit"
// )

// func (suite *ThemeTestSuite) TestGenerateThemeClients() {
//	configPath = goodEnvirontmentPath
//	clients, _, err := generateThemeClients()
//	assert.Nil(suite.T(), err)
//	assert.Equal(suite.T(), 1, len(clients))

//	environments = stringArgArray{[]string{"nope"}}
//	clients, _, err = generateThemeClients()
//	assert.NotNil(suite.T(), err)
//	environments = stringArgArray{[]string{"development"}}

//	allenvs = true
//	clients, _, err = generateThemeClients()
//	assert.Nil(suite.T(), err)
//	assert.Equal(suite.T(), 3, len(clients))
//	allenvs = false

//	configPath = badEnvirontmentPath
//	_, _, err = generateThemeClients()
//	assert.NotNil(suite.T(), err)
// }

// func (suite *ThemeTestSuite) TestShouldUseEnvironment() {
//	environments = stringArgArray{}
//	assert.True(suite.T(), shouldUseEnvironment("development"))

//	environments = stringArgArray{[]string{"production"}}
//	assert.True(suite.T(), shouldUseEnvironment("production"))

//	allenvs = true
//	environments = stringArgArray{}
//	assert.True(suite.T(), shouldUseEnvironment("nope"))
//	allenvs = false

//	environments = stringArgArray{[]string{"p*"}}
//	assert.True(suite.T(), shouldUseEnvironment("production"))
//	assert.True(suite.T(), shouldUseEnvironment("prod"))
//	assert.True(suite.T(), shouldUseEnvironment("puddle"))
//	assert.False(suite.T(), shouldUseEnvironment("development"))

//	environments = stringArgArray{}
//	assert.False(suite.T(), shouldUseEnvironment("production"))
// }

// func (suite *ThemeTestSuite) TestForEachClient() {
//	configPath = goodEnvirontmentPath
//	allenvs = true
//	runtimes := make(chan int, 100)
//	callbacks := make(chan int, 100)
//	runner := forEachClient(func(client kit.ThemeClient, filenames []string, wg *sync.WaitGroup) {
//		defer wg.Done()
//		runtimes <- 1
//	}, func(client kit.ThemeClient, filenames []string, wg *sync.WaitGroup) {
//		callbacks <- 1
//	})
//	assert.NotNil(suite.T(), runner)

//	err := runner(&cobra.Command{}, []string{})
//	assert.Nil(suite.T(), err)
//	assert.Equal(suite.T(), 3, len(runtimes))
//	assert.Equal(suite.T(), 3, len(callbacks))

//	configPath = badEnvirontmentPath
//	err = runner(&cobra.Command{}, []string{})
//	assert.NotNil(suite.T(), err)

//	allenvs = false
// }

// func (suite *ThemeTestSuite) TestSetFlagConfig() {
//	flagConfig.Password = "foo"
//	flagConfig.Domain = "bar"
//	flagConfig.Directory = "my/dir/now"
//	setFlagConfig()

//	config, _ := kit.NewConfiguration()
//	assert.Equal(suite.T(), "foo", config.Password)
//	assert.Equal(suite.T(), "bar", config.Domain)
//	assert.Equal(suite.T(), "my/dir/now", config.Directory)
// }
