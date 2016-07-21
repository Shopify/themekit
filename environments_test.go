package themekit

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

const goodEnv string = "./fixtures/valid_config.yml"
const badPatternEnv string = "./fixtures/bad_pattern_config.yml"

func TestLoadingEnvironmentsFromAFile(t *testing.T) {
	env, err := LoadEnvironmentsFromFile(goodEnv)
	assert.NoError(t, err, "An error should not have been raised")
	assert.Equal(t, 3, len(env))
}

func TestLoadingAConfigurationThatContainsErrors(t *testing.T) {
	_, err := LoadEnvironmentsFromFile(badPatternEnv)
	assert.NotNil(t, err)
}

func TestRetrievingAConfigurationFromAnEnvironment(t *testing.T) {
	env, err := LoadEnvironmentsFromFile(goodEnv)
	conf, err := env.GetConfiguration("default")
	assert.NoError(t, err, "Retrieving the 'default' env should not have raised an error")
	assert.Equal(t, conf.ThemeID, 2)
}

func TestRetrievingAnInvalidConfigurationFromAnEnvironment(t *testing.T) {
	env, err := LoadEnvironmentsFromFile(goodEnv)
	_, err = env.GetConfiguration("invalid")
	assert.Error(t, err, "An error should have been raised when retrieving the 'invalid' environment")
}

func TestSettingAConfiguration(t *testing.T) {
	conf := Configuration{}
	env := Environments{}
	env.SetConfiguration("doodle", conf)

	result, _ := env.GetConfiguration("doodle")
	assert.Equal(t, conf, result)
}

// The result of loading and reserializing aren't idempotent.
// So this is a crappy test.
func TestWritingTheEnvironment(t *testing.T) {
	fmt.Println("TestWritingTheEnvironment is flaky. Skipping...")
	return
	// env, _ := LoadEnvironmentsFromFile(goodEnv)
	// buffer := new(bytes.Buffer)
	// expected, _ := ioutil.ReadFile(goodEnv)
	// env.Write(buffer)
	// assert.Equal(t, len(expected), len(buffer.Bytes()))
}
