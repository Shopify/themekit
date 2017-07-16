package cmd

import (
	"runtime"

	"github.com/Shopify/themekit/kit"

	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of Theme Kit",
	Long:  `All software has versions. This is Theme Kit's version.`,
	Run: func(cmd *cobra.Command, args []string) {
		stdOut.Printf(
			"ThemeKit %s %s/%s",
			kit.ThemeKitVersion.String(),
			runtime.GOOS,
			runtime.GOARCH,
		)
	},
}
