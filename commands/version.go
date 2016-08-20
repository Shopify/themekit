package commands

import (
	"fmt"

	"github.com/Shopify/themekit"
)

// VersionCommand ...
func VersionCommand(args Args, done chan bool) {
	fmt.Println("Theme Kit", themekit.ThemeKitVersion)
	close(done)
}
