package cmd

import (
	"os"

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
		setFlagConfig()
		config, err := kit.NewConfiguration()
		if err != nil {
			return err
		}
		return saveConfiguration(config)
	},
}

func saveConfiguration(config kit.Configuration) error {
	env, err := kit.LoadEnvironments(configPath)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	env.SetConfiguration(environment, config)
	return env.Save(configPath)
}
