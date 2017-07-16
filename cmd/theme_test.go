package cmd

import (
	"testing"

	"github.com/Shopify/themekit/kit"
	"github.com/Shopify/themekit/kittest"
)

func TestThemePreRun(t *testing.T) {
	server := kittest.NewTestServer()
	defer server.Close()
	defer resetArbiter()

	kit.ThemeKitReleasesURL = server.URL + "/themekit_update"

	// just making sure that it does not throw
	ThemeCmd.PersistentPreRun(nil, []string{})
}

func TestThemePostRun(t *testing.T) {
	defer resetArbiter()
	// just making sure that it does not throw
	ThemeCmd.PersistentPostRun(nil, []string{})
}
