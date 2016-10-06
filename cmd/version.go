package cmd

import (
	"fmt"

	"github.com/Shopify/themekit/kit"
)

// VersionCommand ...
func VersionCommand(args Args, done chan bool) {
	fmt.Println("Theme Kit", kit.ThemeKitVersion)
	close(done)
}
