package commands

import (
	"fmt"

	"github.com/Shopify/themekit"
)

func VersionCommand(args Args) chan bool {
	fmt.Println("Theme Kit", themekit.ThemeKitVersion)
	done := make(chan bool)
	close(done)
	return done
}
