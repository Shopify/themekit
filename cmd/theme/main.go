package main

import (
	"github.com/Shopify/themekit/cmd"
	"github.com/Shopify/themekit/kit"
)

func main() {
	if err := cmd.ThemeCmd.Execute(); err != nil {
		kit.LogFatal(err)
	}
}
