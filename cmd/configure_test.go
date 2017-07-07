package cmd

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfigure(t *testing.T) {
	defer os.Remove("config.yml")
	defer resetArbiter()
	arbiter.configPath = "config.yml"

	err := configureCmd.RunE(nil, []string{})
	assert.NotNil(t, err)

	arbiter.flagConfig.Password = "foo"
	arbiter.flagConfig.Domain = "myshop.myshopify.com"
	arbiter.flagConfig.ThemeID = "1"
	arbiter.setFlagConfig()

	err = configureCmd.RunE(nil, []string{})
	assert.Nil(t, err)

	arbiter.configPath = "does_not_exist/nope.xm"
	err = configureCmd.RunE(nil, []string{})
	assert.NotNil(t, err)
}
