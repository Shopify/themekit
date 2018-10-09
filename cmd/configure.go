package cmd

import (
	"github.com/spf13/cobra"

	"github.com/Shopify/themekit/src/cmdutil"
)

var configureCmd = &cobra.Command{
	Use:   "configure",
	Short: "Create a configuration file",
	Long: `Configure will create a new configuration file to
 access shopify using the theme kit.

 For more documentation please see http://shopify.github.io/themekit/commands/#configure
 `,
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmdutil.ForDefaultClient(flags, args, createConfig)
	},
}

func createConfig(ctx *cmdutil.Ctx) error {
	for _, name := range ctx.Flags.Environments {
		if _, err := ctx.Conf.Set(name, *ctx.Env); err != nil {
			return err
		}
	}
	return ctx.Conf.Save()
}
