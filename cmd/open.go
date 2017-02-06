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
	previewURL := fmt.Sprintf("https://%s?preview_theme_id=%s",
		client.Config.Domain,
		client.Config.ThemeID)

	if verbose {
		kit.Printf("[%s] opening %s",
			kit.GreenText(client.Config.Environment),
			kit.GreenText(previewURL))
	}

	err := open.Run(previewURL)
	if err != nil {
		kit.LogErrorf("[%s] %s",
			kit.GreenText(client.Config.Environment),
			kit.RedText(err))
	}
}
