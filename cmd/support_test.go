package cmd

import (
	"github.com/Shopify/themekit/kit"
)

func init() {
	resetArbiter()
}

func resetArbiter() {
	arbiter = newCommandArbiter()
	arbiter.verbose = true
	arbiter.setFlagConfig()
}

func getClient() (kit.ThemeClient, error) {
	if err := arbiter.generateThemeClients(nil, []string{}); err != nil {
		return kit.ThemeClient{}, err
	}
	return arbiter.activeThemeClients[0], nil
}
