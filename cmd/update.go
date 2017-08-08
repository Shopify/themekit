package cmd

import (
	"github.com/spf13/cobra"

	"github.com/Shopify/themekit/kit"
)

const afterUpdateMessage = `
Successfully updated to theme kit version %v,
If you have troubles with this release please
report them to https://github.com/Shopify/themekit/issues
If your troubles are preventing you from working
you can roll back to the previous version using
the command 'theme update --version=v%s'
`

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update Theme kit to the newest version.",
	Long: `Update will check for a new release, then
if there is an applicable update it will download it and apply it.

For more documentation please see http://shopify.github.io/themekit/commands/#update
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		stdOut.Printf(
			"Updating from %s to %s",
			yellow(kit.ThemeKitVersion),
			yellow(updateVersion),
		)

		err := kit.InstallThemeKitVersion(updateVersion)

		if err == nil {
			stdOut.Printf(afterUpdateMessage, green(updateVersion), yellow(kit.ThemeKitVersion))
		}

		return err
	},
}
