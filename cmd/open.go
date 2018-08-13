package cmd

import (
	"fmt"

	"github.com/skratchdot/open-golang/open"
	"github.com/spf13/cobra"

	"github.com/Shopify/themekit/src/cmdutil"
	"github.com/Shopify/themekit/src/colors"
)

var openCmd = &cobra.Command{
	Use:   "open",
	Short: "Open the preview for your store.",
	Long: `Open will open the preview page in your browser as well as print out
url for your reference`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmdutil.ForSingleClient(flags, args, func(ctx cmdutil.Ctx) error {
			return preview(ctx, func(url string) error {
				if ctx.Flags.With == "" {
					return open.Run(url)
				}
				return open.RunWith(url, ctx.Flags.With)
			})
		})
	},
}

func preview(ctx cmdutil.Ctx, openFunc func(string) error) error {
	url := fmt.Sprintf("https://%s?preview_theme_id=%s", ctx.Env.Domain, ctx.Env.ThemeID)
	if ctx.Flags.Edit {
		url = fmt.Sprintf("https://%s/admin/themes/%s/editor", ctx.Env.Domain, ctx.Env.ThemeID)
	}
	ctx.Log.Printf("[%s] opening %s", colors.Green(ctx.Env.Name), colors.Green(url))
	if err := openFunc(url); err != nil {
		return fmt.Errorf("[%s] Error opening: %s", colors.Green(ctx.Env.Name), colors.Red(err))
	}
	return nil
}
