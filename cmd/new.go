package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/Shopify/themekit/src/cmdutil"
	"github.com/Shopify/themekit/src/colors"
	"github.com/Shopify/themekit/src/timber"
)

var bootstrapCmd = &cobra.Command{
	Use:   "bootstrap",
	Short: "[DEPRECATED] please use `new` instead",
	Long: `[DEPRECATED] Bootstrap has been deprecated and renames to new. New will
	work with all the same arguments as bootstrap.
  `,
	RunE: func(cmd *cobra.Command, args []string) error {
		return fmt.Errorf("bootstrap has been deprecated please use `new` instead")
	},
}

var newCmd = &cobra.Command{
	Use:   "new",
	Short: "New will create theme using Shopify Timber",
	Long: `New will download the latest release of Timber,
  The most popular theme on Shopify. New will also setup
  your config file and create a new theme id for you.

  For more documentation please see http://shopify.github.io/themekit/commands/#new
  `,
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmdutil.ForDefaultClient(flags, args, func(ctx cmdutil.Ctx) error {
			url, name, err := getNewThemeDetails(flags, timber.GetVersionPath)
			if err != nil {
				return err
			}
			return newTheme(ctx, name, url)
		})
	},
}

func newTheme(ctx cmdutil.Ctx, name, url string) error {
	ctx.Log.Printf("[%s] creating new theme \"%s\" from %s", colors.Yellow(ctx.Env.Domain), colors.Yellow(name), colors.Yellow(url))

	theme, err := ctx.Client.CreateNewTheme(name, url)
	if err != nil {
		return err
	}

	ctx.Log.Printf("[%s] created theme", colors.Yellow(ctx.Env.Domain))

	ctx.Flags.ThemeID = fmt.Sprintf("%v", theme.ID)
	if err := createConfig(ctx); err != nil {
		return err
	}

	ctx.Log.Printf("[%s] created config", colors.Yellow(ctx.Env.Domain))

	for {
		if theme, err := ctx.Client.GetInfo(); err != nil {
			ctx.ErrLog.Println("Encountered an error while checking new theme. Please run `theme download` to complete the setup.")
			return err
		} else if theme.Previewable {
			ctx.Log.Println("downloading...")
			break
		}
		ctx.Log.Println("processing...")
		time.Sleep(500 * time.Millisecond)
	}

	return download(ctx)
}

func getNewThemeDetails(flags cmdutil.Flags, getVer func(string) (string, error)) (name, url string, err error) {
	name, url = flags.Name, flags.URL

	if name == "" && url != "" {
		parts := strings.Split(flags.URL, "/")
		name = flags.Prefix + strings.Replace(parts[len(parts)-1], ".zip", "", 1)
	} else if name == "" {
		name = flags.Prefix + "Timber-" + flags.Version
	}

	if url == "" {
		url, err = getVer(flags.Version)
	}

	return name, url, err
}
