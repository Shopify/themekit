package commands

import (
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
)

const badPatternEnv string = "../fixtures/bad_pattern_config.yml"

func TestMigrateDoesNotModifyWhenYAMLIsInvalid(t *testing.T) {
	expected, _ := ioutil.ReadFile(badPatternEnv)
	err := MigrateConfiguration(badPatternEnv)
	actual, _ := ioutil.ReadFile(badPatternEnv)
	assert.Error(t, err, "An error should've been returned")
	assert.Equal(t, expected, actual)
}
