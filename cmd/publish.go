package cmd

import (
	"github.com/spf13/cobra"

	"github.com/Shopify/themekit/src/cmdutil"
)

var publishCmd = &cobra.Command{
	Use:   "publish",
	Short: "publish a theme",
	Long: `Publish will update the theme to be your current publish theme. Select
the theme you want to publish using the env flag.

 For more documentation please see http://shopify.github.io/themekit/commands/#deploy
 `,
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmdutil.ForSingleClient(flags, args, publish)
	},
}

func publish(ctx *cmdutil.Ctx) error {
	return ctx.Client.PublishTheme()
}
