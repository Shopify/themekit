package cmd

import "github.com/spf13/cobra"

var replaceCmd = &cobra.Command{
	Use:   "replace <filenames>",
	Short: "Overwrite theme file(s)",
	Long: `Replace will overwrite specific files if provided with file names.
If replace is not provided with file names then it will replace all
the files on shopify with your local files. Any files that do not
exist on your local machine will be removed from shopify.

For more documentation please see http://shopify.github.io/themekit/commands/#replace
`,
	PreRunE:  arbiter.generateThemeClients,
	RunE:     arbiter.forEachClient(deploy(true)),
	PostRunE: arbiter.forEachClient(uploadSettingsData),
}
