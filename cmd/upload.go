package cmd

import (
	"sync"

	"github.com/spf13/cobra"

	"github.com/Shopify/themekit/kit"
)

var uploadCmd = &cobra.Command{
	Use:   "upload <filenames>",
	Short: "Upload theme file(s) to shopify",
	Long: `Upload will upload specific files to shopify servers if provided file names.
If no filenames are provided then upload will upload every file in the project
to shopify.`,
	RunE: forEachClient(upload),
}

func upload(client kit.ThemeClient, filenames []string, wg *sync.WaitGroup) {
	defer wg.Done()
	if len(filenames) == 0 {
		localAssets, err := client.LocalAssets()
		if err != nil {
			kit.LogError(err)
			return
		}

		for _, asset := range localAssets {
			wg.Add(1)
			go performUpload(client, asset, wg)
		}
	} else {
		for _, filename := range filenames {
			asset, err := client.LocalAsset(filename)
			if err != nil {
				kit.LogError(err)
				return
			}
			wg.Add(1)
			go performUpload(client, asset, wg)
		}
	}
}

func performUpload(client kit.ThemeClient, asset kit.Asset, wg *sync.WaitGroup) {
	resp, err := client.UpdateAsset(asset)
	if err != nil {
		kit.Print(err)
	} else {
		kit.Printf(
			"Successfully performed Update on file %s from %s",
			kit.GreenText(asset.Key),
			kit.YellowText(resp.Host),
		)
	}
	wg.Done()
}
