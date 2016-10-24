package cmd

import (
	"github.com/spf13/cobra"

	"github.com/Shopify/themekit/kit"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update Theme kit to the newest verion.",
	Long: `Update will check for a new release, then
if there is an applicable update it will
download it and apply it.`,
	RunE: doUpdate,
}

func doUpdate(cmd *cobra.Command, args []string) error {
	return kit.InstallThemeKitVersion(updateVersion)
}
