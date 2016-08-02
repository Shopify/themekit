package commands

import (
	"fmt"

	"github.com/Shopify/themekit"
)

// VersionCommand ...
func VersionCommand(args Args) chan bool {
	fmt.Println("Theme Kit", themekit.ThemeKitVersion)
	done := make(chan bool)
	close(done)
	return done
}
