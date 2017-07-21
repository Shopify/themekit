package cmd

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Shopify/themekit/kit"
	"github.com/Shopify/themekit/kittest"
)

func init() {
	resetArbiter()
}

func resetArbiter() {
	arbiter = newCommandArbiter()
	arbiter.verbose = true
	arbiter.setFlagConfig()
	resetLog()
}

func getClient() (kit.ThemeClient, error) {
	if err := arbiter.generateThemeClients(nil, []string{}); err != nil {
		return kit.ThemeClient{}, err
	}
	return arbiter.activeThemeClients[0], nil
}

func TestNewCommandArbiter(t *testing.T) {
	arb := newCommandArbiter()
	assert.NotNil(t, arb)
	assert.NotNil(t, arb.progress)
	assert.NotNil(t, arb.flagConfig)
	assert.False(t, arb.configPath == "")
}

func TestGenerateManifest(t *testing.T) {
	defer resetArbiter()
	assert.Nil(t, arbiter.generateManifest())
	assert.NotNil(t, arbiter.manifest)
}

func TestGenerateThemeClients(t *testing.T) {
	server := kittest.NewTestServer()
	defer server.Close()

	err := arbiter.generateThemeClients(nil, []string{})
	assert.NotNil(t, err)
	assert.True(t, strings.Contains(err.Error(), "Could not find config file"))

	assert.Nil(t, kittest.GenerateConfig(server.URL, false))

	err = arbiter.generateThemeClients(nil, []string{})
	assert.NotNil(t, err)
	assert.True(t, strings.Contains(err.Error(), "Invalid yaml found"))

	assert.Nil(t, kittest.GenerateConfig(server.URL, true))
	defer kittest.Cleanup()
	defer resetArbiter()

	arbiter.environments = stringArgArray{[]string{"nope"}}
	err = arbiter.generateThemeClients(nil, []string{})
	assert.NotNil(t, err)
	assert.True(t, strings.Contains(err.Error(), "Could not load any valid environments"))

	arbiter.environments = stringArgArray{[]string{"production"}}
	arbiter.disableIgnore = true
	assert.Nil(t, arbiter.generateThemeClients(nil, []string{}))
	assert.Equal(t, 1, len(arbiter.activeThemeClients))
	assert.Equal(t, 3, len(arbiter.allThemeClients))
	assert.Equal(t, 0, len(arbiter.allThemeClients[0].Config.IgnoredFiles))

	arbiter.environments = stringArgArray{[]string{}}
	assert.Nil(t, arbiter.generateThemeClients(nil, []string{}))

	kittest.Cleanup()
	assert.Nil(t, kittest.GenerateProxyConfig(server.URL, false))
	arbiter.generateThemeClients(nil, []string{})
	assert.True(t, strings.Contains(stdOutOutput.String(), "Proxy URL detected from Configuration"))
}

func TestShouldUseEnvironment(t *testing.T) {
	defer resetArbiter()

	arbiter.environments = stringArgArray{}
	assert.True(t, arbiter.shouldUseEnvironment("development"))
	assert.False(t, arbiter.shouldUseEnvironment("production"))

	arbiter.environments = stringArgArray{[]string{"production"}}
	assert.True(t, arbiter.shouldUseEnvironment("production"))
	assert.False(t, arbiter.shouldUseEnvironment("development"))

	arbiter.allenvs = true
	arbiter.environments = stringArgArray{}
	assert.True(t, arbiter.shouldUseEnvironment("nope"))
	arbiter.allenvs = false

	arbiter.environments = stringArgArray{[]string{"p*", "other"}}
	assert.True(t, arbiter.shouldUseEnvironment("production"))
	assert.True(t, arbiter.shouldUseEnvironment("prod"))
	assert.True(t, arbiter.shouldUseEnvironment("puddle"))
	assert.False(t, arbiter.shouldUseEnvironment("development"))
	assert.True(t, arbiter.shouldUseEnvironment("other"))
}

func TestForEachClient(t *testing.T) {
	server := kittest.NewTestServer()
	defer server.Close()
	assert.Nil(t, kittest.GenerateConfig(server.URL, true))
	defer kittest.Cleanup()
	defer resetArbiter()

	arbiter.allenvs = true
	_, err := getClient()
	if assert.Nil(t, err) {
		runner := arbiter.forEachClient(func(client kit.ThemeClient, filenames []string) error {
			return nil
		})
		assert.Nil(t, runner(nil, []string{}))
	}
}

func TestForSingleClient(t *testing.T) {
	server := kittest.NewTestServer()
	defer server.Close()
	assert.Nil(t, kittest.GenerateConfig(server.URL, true))
	defer kittest.Cleanup()
	defer resetArbiter()

	arbiter.allenvs = true
	_, err := getClient()
	if assert.Nil(t, err) {
		runner := arbiter.forSingleClient(func(client kit.ThemeClient, filenames []string) error {
			return nil
		})
		err := runner(nil, []string{})
		assert.NotNil(t, err)
		assert.True(t, strings.Contains(err.Error(), "more than one env"))
	}

	arbiter.allenvs = false
	_, err = getClient()
	if assert.Nil(t, err) {
		runner := arbiter.forSingleClient(func(client kit.ThemeClient, filenames []string) error {
			return nil
		})
		assert.Nil(t, runner(nil, []string{}))
	}
}

func TestSetFlagConfig(t *testing.T) {
	defer resetArbiter()
	arbiter.flagConfig.Password = "foo"
	arbiter.flagConfig.Domain = "bar.myshopify.com"
	arbiter.flagConfig.Directory = "my/dir/now"
	arbiter.flagConfig.ThemeID = "123"
	arbiter.setFlagConfig()

	config, err := kit.NewConfiguration()
	assert.Nil(t, err)
	assert.Equal(t, "foo", config.Password)
	assert.Equal(t, "bar.myshopify.com", config.Domain)
	assert.Equal(t, "my/dir/now", config.Directory)
	assert.Equal(t, "123", config.ThemeID)
}

func TestNewProgressBar(t *testing.T) {
	defer resetArbiter()
	arbiter.verbose = false
	bar := arbiter.newProgressBar(1, "Dev")
	assert.NotNil(t, bar)
	arbiter.verbose = true
	bar = arbiter.newProgressBar(1, "Dev")
	assert.Nil(t, bar)
	assert.Equal(t, 1, arbiter.progress.BarCount())
}

func TestGenerateAssetActions(t *testing.T) {
	server := kittest.NewTestServer()
	defer server.Close()
	assert.Nil(t, kittest.GenerateConfig(server.URL, true))
	assert.Nil(t, kittest.GenerateProject())
	defer kittest.Cleanup()
	defer resetArbiter()

	client, err := getClient()
	if assert.Nil(t, err) {
		actions, err := arbiter.generateAssetActions(client, []string{}, true)
		assert.Nil(t, err)
		assert.Equal(t, len(kittest.ProjectFiles)-1+2, len(actions)) //remove .gitkeep add 2 removes

		server.Close()
		_, err = arbiter.generateAssetActions(client, []string{}, true)
		assert.NotNil(t, err)
	}
}

func TestPreflightCheck(t *testing.T) {
	server := kittest.NewTestServer()
	defer server.Close()
	assert.Nil(t, kittest.GenerateConfig(server.URL, true))
	defer kittest.Cleanup()
	defer resetArbiter()

	_, err := getClient()
	if assert.Nil(t, err) {
		assert.Nil(t, arbiter.preflightCheck(map[string]assetAction{}, true))

		err := arbiter.preflightCheck(map[string]assetAction{
			"assets/hello.txt": {event: kit.Remove},
		}, true)

		assert.NotNil(t, err)
		assert.True(t, strings.Contains(err.Error(), "Diff"))

		assert.Nil(t, arbiter.preflightCheck(map[string]assetAction{
			"assets/hello.txt": {event: kit.Update},
		}, false))

		arbiter.force = true
		assert.Nil(t, arbiter.preflightCheck(map[string]assetAction{
			"assets/hello.txt": {event: kit.Update},
		}, true))
	}
}
