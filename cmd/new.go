package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	_ "github.com/Shopify/themekit/cmd/static" // This will import the asset bundle
	"github.com/Shopify/themekit/src/cmdutil"
	"github.com/Shopify/themekit/src/colors"
	"github.com/Shopify/themekit/src/shopify"
	"github.com/Shopify/themekit/src/static"
)

var newCmd = &cobra.Command{
	Use:   "new",
	Short: "New will create theme using Shopify Timber",
	Long: `New will download the latest release of Timber,
  The most popular theme on Shopify. New will also setup
  your config file and create a new theme id for you.

  For more documentation please see http://shopify.github.io/themekit/commands/#new
  `,
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmdutil.ForDefaultClient(flags, args, func(ctx *cmdutil.Ctx) error {
			return newTheme(ctx, static.Unbundle)
		})
	},
}

func newTheme(ctx *cmdutil.Ctx, generate func(ctx *cmdutil.Ctx) error) error {
	theme, err := ctx.Client.CreateNewTheme(ctx.Flags.Name)
	if err != nil {
		if err == shopify.ErrThemeNameRequired {
			return fmt.Errorf("a theme name is required, please use the --name flag to define it")
		}
		return err
	}
	ctx.Log.Printf("[%s] theme created", colors.Yellow(ctx.Env.Domain))

	ctx.Env.ThemeID = fmt.Sprintf("%v", theme.ID)
	if err := createConfig(ctx); err != nil {
		return err
	}
	ctx.Log.Printf("[%s] config created", colors.Yellow(ctx.Env.Domain))

	if err := generate(ctx); err != nil {
		return err
	}

	ctx.Log.Printf("[%s] uploading new files to shopify", colors.Yellow(ctx.Env.Domain))
	return deploy(ctx)
}
