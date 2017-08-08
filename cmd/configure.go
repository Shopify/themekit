package cmd

import (
	"github.com/spf13/cobra"

	"github.com/Shopify/themekit/kit"
)

var configureCmd = &cobra.Command{
	Use:   "configure",
	Short: "Create a configuration file",
	Long: `Configure will create a new configuration file to
access shopify using the theme kit.

For more documentation please see http://shopify.github.io/themekit/commands/#configure
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		config, err := kit.NewConfiguration()
		if err != nil {
			return err
		}
		return saveConfiguration(config)
	},
}
