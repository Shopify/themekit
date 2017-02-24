package cmd

import (
	"fmt"
	"sync"

	"github.com/skratchdot/open-golang/open"
	"github.com/spf13/cobra"

	"github.com/Shopify/themekit/kit"
)

var customizeCmd = &cobra.Command{
	Use:   "customize",
	Short: "Open the theme customizer for your store.",
	Long: `Open will open the customizer in your browser as well as print out
url for your reference`,
	RunE: forEachClient(preview),
}

func preview(client kit.ThemeClient, filenames []string, wg *sync.WaitGroup) {
	defer wg.Done()
	previewURL := fmt.Sprintf("https://%sadmin/themes/%s/editor",
		client.Config.Domain,
		client.Config.ThemeID)

	kit.Printf("[%s] opening %s",
		kit.GreenText(client.Config.Environment),
		kit.GreenText(previewURL))

	err := open.Run(previewURL)
	if err != nil {
		kit.LogErrorf("[%s] %s",
			kit.GreenText(client.Config.Environment),
			kit.RedText(err))
	}
}
