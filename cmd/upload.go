package cmd

import (
	"sync"

	"github.com/spf13/cobra"

	"github.com/Shopify/themekit/kit"
	"github.com/Shopify/themekit/theme"
)

var uploadCmd = &cobra.Command{
	Use:   "upload <filenames>",
	Short: "Upload theme file(s) to shopify",
	Long: `Upload will upload specific files to shopify servers if provided file names.
If no filenames are provided then upload will upload every file in the project
to shopify.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := initializeConfig(cmd.Name(), true); err != nil {
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
			performUpload(client, asset, wg)
		}
	} else {
		for _, filename := range filenames {
			asset, err := client.LocalAsset(filename)
			if err != nil {
				return err
			} else {
				performUpload(client, asset, wg)
			}
		}
	}
	wg.Done()
	return nil
}

func performUpload(client kit.ThemeClient, asset theme.Asset, wg *sync.WaitGroup) {
	wg.Add(1)
	client.UpdateAsset(asset, func(resp *kit.ShopifyResponse, err kit.Error) {
		if err != nil {
			kit.Logf(err.Error())
			kit.Logf(resp.String())
		} else {
			kit.Logf(
				"Successfully performed %s on file %s from %s",
				kit.BlueText(resp.EventType),
				kit.BlueText(resp.Asset.Key),
				kit.YellowText(resp.Host),
			)
		}
		wg.Done()
	})
}
