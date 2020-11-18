package cmd

import (
	"fmt"

	"github.com/skratchdot/open-golang/open"
	"github.com/spf13/cobra"

	"github.com/Shopify/themekit/src/cmdutil"
	"github.com/Shopify/themekit/src/colors"
)

type runWithFunc func(url, prog string) error
type runFunc func(url string) error

var openCmd = &cobra.Command{
	Use:   "open",
	Short: "Open the preview for your store.",
	Long: `Open will open the preview page in your browser as well as print out
url for your reference`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// open should not care about the live theme
		flags.AllowLive = true
		return cmdutil.ForSingleClient(flags, args, func(ctx *cmdutil.Ctx) error {
			return preview(ctx, open.Run, open.RunWith)
		})
	},
}

func preview(ctx *cmdutil.Ctx, run runFunc, runWith runWithFunc) error {
	url := fmt.Sprintf("https://%s?preview_theme_id=%s", ctx.Env.Domain, ctx.Env.ThemeID)
	if ctx.Flags.HidePreviewBar {
		url += "&pb=0"
	}
	if ctx.Flags.Edit {
		url = fmt.Sprintf("https://%s/admin/themes/%s/editor", ctx.Env.Domain, ctx.Env.ThemeID)
	}
	ctx.Log.Printf("[%s] opening %s", colors.Green(ctx.Env.Name), colors.Green(url))

	if ctx.Flags.With == "" {
		if err := run(url); err != nil {
			return fmt.Errorf("[%s] Error opening: %s", colors.Green(ctx.Env.Name), colors.Red(err))
		}
	} else if err := runWith(url, ctx.Flags.With); err != nil {
		return fmt.Errorf("[%s] Error opening: %s", colors.Green(ctx.Env.Name), colors.Red(err))
	}

	return nil
}
