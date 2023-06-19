package cmd

import (
	"github.com/spf13/cobra"

	"github.com/Shopify/themekit/src/cmdutil"
	"github.com/Shopify/themekit/src/colors"
)

var publishCmd = &cobra.Command{
	Use:   "publish",
	Short: "publish a theme",
	Long: `Publish will update the theme to be your current publish theme. Select
the theme you want to publish using the env flag.

 For more information, refer to https://shopify.dev/tools/theme-kit/command-reference#deploy.
 `,
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmdutil.ForSingleClient(flags, args, publish)
	},
}

func publish(ctx *cmdutil.Ctx) (err error) {
	if err = ctx.Client.PublishTheme(); err == nil {
		ctx.Log.Printf("[%s] Successfully published theme %s", colors.Green(ctx.Env.Name), colors.Green(ctx.Env.ThemeID))
	}
	return err
}
