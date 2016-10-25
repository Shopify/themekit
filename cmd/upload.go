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
	RunE: func(cmd *cobra.Command, args []string) error {
		themeClients, err := generateThemeClients()
		if err != nil {
			return err
		}

		wg := sync.WaitGroup{}
		for _, client := range themeClients {
			wg.Add(1)
			go upload(client, args, &wg)
		}
		wg.Wait()
		return nil
	},
}

func upload(client kit.ThemeClient, filenames []string, wg *sync.WaitGroup) error {
	if len(filenames) == 0 {
		localAssets, err := client.LocalAssets()
		if err != nil {
			return err
		}

		for _, asset := range localAssets {
			wg.Add(1)
			go performUpload(client, asset, wg)
		}
	} else {
		for _, filename := range filenames {
			asset, err := client.LocalAsset(filename)
			if err != nil {
				return err
			}
			wg.Add(1)
			go performUpload(client, asset, wg)
		}
	}
	wg.Done()
	return nil
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
