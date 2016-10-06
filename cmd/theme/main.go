package main

import (
	"fmt"
	"os"

	"github.com/Shopify/themekit/cmd"
)

func main() {
	if err := cmd.ThemeCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(-1)
	}
}
