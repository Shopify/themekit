package commands

import (
	"fmt"
	"github.com/Shopify/themekit"
)

func VersionCommand(args map[string]interface{}) chan bool {
	fmt.Println("Theme Kit", themekit.ThemeKitVersion)
	res := make(chan bool)
	close(res)
	return res
}
