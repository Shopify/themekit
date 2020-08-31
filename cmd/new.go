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
	Long: `New will create a new theme on Shopify, generate a minimal template of
	all the required files that a theme needs to be functional and then setup
  your config file for working with your new theme.

  For more documentation please see http://shopify.github.io/themekit/commands/#new
  `,
	RunE: func(cmd *cobra.Command, args []string) error {
		// This is a hack to get around theme ID validation for the list operation which doesnt need it
		flags.ThemeID = "1337"
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
