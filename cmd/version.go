package cmd

import (
	"fmt"
	"github.com/Shopify/themekit/kit"

	"github.com/spf13/cobra"
)

const versionString = `
----------------------------------------------------------------
		%s ThemeKit - Shopify Theme Tool
----------------------------------------------------------------
`

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of Theme Kit",
	Long:  `All software has versions. This is Theme Kit's version.`,
	Run: func(cmd *cobra.Command, args []string) {
		stdOut.Print(green(fmt.Sprintf(versionString, kit.ThemeKitVersion.String())))
	},
}
