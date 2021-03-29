package cmd

import (
	"bytes"
	"errors"
	"fmt"
	"strconv"
	"text/template"

	"github.com/spf13/cobra"

	"github.com/Shopify/themekit/src/cmdutil"
	"github.com/Shopify/themekit/src/shopify"
)

var (
	errNoThemes     = errors.New("No available themes")
	availableThemes = template.Must(template.New("availableThemes").Parse(`Available theme versions:
  {{- range . }}
  [{{ .ID }}]{{if eq .Role "main"}}[live]{{ end }} {{ .Name }}
  {{- end }}`))
)

var getCmd = &cobra.Command{
	Use:   "get",
	Short: "Get a theme and config from shopify",
	Long: `Get will get a theme from shopify and create a config.yml for the theme
 that it accesses. To get a list of all themes that you can access, you can pass in
 a --list flag.

 For more information, refer to https://shopify.dev/tools/theme-kit/command-reference#get.
 `,
	RunE: func(cmd *cobra.Command, args []string) error {
		// get should not care about the live theme
		flags.AllowLive = true
		if flags.List {
			return listThemes(flags, args)
		} else if flags.Live {
			theme, err := getLiveTheme(flags, args)
			if err != nil {
				return err
			}
			flags.ThemeID = strconv.Itoa(int(theme.ID))
		}
		return cmdutil.ForDefaultClient(flags, args, getTheme)
	},
}

func getTheme(ctx *cmdutil.Ctx) error {
	if err := createConfig(ctx); err != nil {
		return err
	}
	return download(ctx)
}

func listThemes(flags cmdutil.Flags, args []string) error {
	flags.ThemeID = "1337"
	themes, err := getDefaultThemes(flags, args)
	if err != nil {
		return err
	}
	return cmdutil.ForDefaultClient(flags, args, func(ctx *cmdutil.Ctx) error {
		var tpl bytes.Buffer
		availableThemes.Execute(&tpl, themes)
		ctx.Log.Println(tpl.String())
		return nil
	})
}

func getLiveTheme(flags cmdutil.Flags, args []string) (shopify.Theme, error) {
	themes, err := getDefaultThemes(flags, args)
	if err != nil {
		return shopify.Theme{}, err
	}
	for _, theme := range themes {
		if theme.Role == "main" {
			return theme, nil
		}
	}
	return shopify.Theme{}, fmt.Errorf("No live theme found")
}

func getDefaultThemes(flags cmdutil.Flags, args []string) ([]shopify.Theme, error) {
	// This is a hack to get around theme ID validation for the list operation which doesnt need it
	flags.ThemeID = "1337"
	var err error
	var themes []shopify.Theme
	return themes, cmdutil.ForDefaultClient(flags, args, func(ctx *cmdutil.Ctx) error {
		if themes, err = ctx.Client.Themes(); err != nil {
			return err
		} else if len(themes) == 0 {
			return errNoThemes
		}
		return nil
	})
}

func withThemes(flags cmdutil.Flags, args []string, fn func(ctx *cmdutil.Ctx, themes []shopify.Theme) error) error {
	// This is a hack to get around theme ID validation for the list operation which doesnt need it
	flags.ThemeID = "1337"
	return cmdutil.ForDefaultClient(flags, args, func(ctx *cmdutil.Ctx) error {
		themes, err := ctx.Client.Themes()
		if err != nil {
			return err
		} else if len(themes) == 0 {
			return errNoThemes
		}
		return fn(ctx, themes)
	})
}
