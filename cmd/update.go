package cmd

import (
	"github.com/spf13/cobra"

	"github.com/Shopify/themekit/kit"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update Theme kit to the newest version.",
	Long: `Update will check for a new release, then
if there is an applicable update it will download it and apply it.

For more documentation please see http://shopify.github.io/themekit/commands/#update
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return kit.InstallThemeKitVersion(updateVersion)
	},
}
