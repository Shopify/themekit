package cmd

import (
	"fmt"
	"sync"

	"github.com/skratchdot/open-golang/open"
	"github.com/spf13/cobra"

	"github.com/Shopify/themekit/kit"
)

var openCmd = &cobra.Command{
	Use:   "open",
	Short: "Open the preview for your store.",
	Long: `Open will open the preview page in your browser as well as print out
url for your reference`,
	RunE: forEachClient(preview),
}

func preview(client kit.ThemeClient, filenames []string, wg *sync.WaitGroup) {
	defer wg.Done()
	url := fmt.Sprintf("https://%s?preview_theme_id=%s",
		client.Config.Domain,
		client.Config.ThemeID)

	if openEdit {
		url = fmt.Sprintf("https://%s/admin/themes/%s",
			client.Config.Domain,
			client.Config.ThemeID)
	}

	openURL(client.Config.Environment, url)
}

func openURL(env, url string) {
	if verbose {
		kit.Printf("[%s] opening %s", kit.GreenText(env), kit.GreenText(url))
	}

	err := open.Run(url)
	if err != nil {
		kit.LogErrorf("[%s] %s", kit.GreenText(env), kit.RedText(err))
	}
}
