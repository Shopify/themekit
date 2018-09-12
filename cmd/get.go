package cmd

import (
	"bytes"
	"errors"
	"text/template"

	"github.com/spf13/cobra"

	"github.com/Shopify/themekit/src/cmdutil"
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

 For more documentation please see http://shopify.github.io/themekit/commands/#get
 `,
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmdutil.ForDefaultClient(flags, args, getTheme)
	},
}

func getTheme(ctx *cmdutil.Ctx) error {
	if ctx.Flags.List {
		themes, err := ctx.Client.Themes()
		if err != nil {
			return err
		} else if len(themes) == 0 {
			return errNoThemes
		}

		var tpl bytes.Buffer
		availableThemes.Execute(&tpl, themes)
		ctx.Log.Println(tpl.String())
		return nil
	}

	if err := createConfig(ctx); err != nil {
		return err
	}

	return download(ctx)
}
