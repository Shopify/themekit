package cmd

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Shopify/themekit/kit"
	"github.com/Shopify/themekit/kittest"
)

func TestUpdate(t *testing.T) {
	server := kittest.NewTestServer()
	defer server.Close()
	defer resetArbiter()

	kit.ThemeKitReleasesURL = server.URL + "/themekit_update"
	updateVersion = "0.0.0"

	expected := fmt.Sprintf(
		"Updating from %s to %s\n",
		yellow(kit.ThemeKitVersion),
		yellow("0.0.0"),
	)

	err := updateCmd.RunE(nil, []string{})
	assert.Equal(t, "version 0.0.0 not found", err.Error())
	assert.Equal(t, expected, stdOutOutput.String())
}
