package cmd

import (
	"fmt"

	"github.com/skratchdot/open-golang/open"
	"github.com/spf13/cobra"

	"github.com/Shopify/themekit/kit"
)

var openFunc = open.Run

var openCmd = &cobra.Command{
	Use:   "open",
	Short: "Open the preview for your store.",
	Long: `Open will open the preview page in your browser as well as print out
url for your reference`,
	PreRunE: arbiter.generateThemeClients,
	RunE:    arbiter.forSingleClient(preview),
}

func preview(client kit.ThemeClient, filenames []string) error {
	themeID := client.Config.ThemeID

	if openEdit && themeID == "live" {
		return fmt.Errorf(
			"[%s] cannot open editor for live theme without theme id",
			green(client.Config.Environment),
		)
	}

	if themeID == "live" {
		stdOut.Printf(
			"[%s] This theme is live so preview is the same as your live shop.",
			green(client.Config.Environment),
		)
		themeID = ""
	}

	url := fmt.Sprintf("https://%s?preview_theme_id=%s", client.Config.Domain, themeID)

	if openEdit {
		url = fmt.Sprintf("https://%s/admin/themes/%s/editor",
			client.Config.Domain,
			client.Config.ThemeID)
	}

	stdOut.Printf(
		"[%s] opening %s",
		green(client.Config.Environment),
		green(url),
	)

	if err := openFunc(url); err != nil {
		return fmt.Errorf(
			"[%s] Error opening: %s",
			green(client.Config.Environment),
			red(err),
		)
	}

	return nil
}
