package cmd

import (
	"strconv"

	"github.com/spf13/cobra"

	"github.com/Shopify/themekit/src/cmdutil"
)

var configureCmd = &cobra.Command{
	Use:   "configure",
	Short: "Create a configuration file",
	Long: `Configure will create a new configuration file to
 access shopify using the theme kit.

 For more information, refer to https://shopify.dev/tools/theme-kit/command-reference#configure.
 `,
	RunE: func(cmd *cobra.Command, args []string) error {
		// get should not care about the live theme
		flags.AllowLive = true
		if flags.Live {
			theme, err := getLiveTheme(flags, args)
			if err != nil {
				return err
			}
			flags.ThemeID = strconv.Itoa(int(theme.ID))
		}
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
