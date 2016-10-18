package cmd

import (
	"fmt"
	"os"
	"sync"

	"github.com/spf13/cobra"

	"github.com/Shopify/themekit/kit"
	"github.com/Shopify/themekit/theme"
)

var removeCmd = &cobra.Command{
	Use:   "remove <filenames>",
	Short: "Remove theme file(s) from shopify",
	Long:  `Remove will delete all specified files from shopify servers.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := initializeConfig(cmd.Name(), true); err != nil {
			return err
		}

		wg := sync.WaitGroup{}
		for _, client := range themeClients {
			wg.Add(1)
			go remove(client, args, &wg)
		}
		wg.Wait()
		return nil
	},
}

func remove(client kit.ThemeClient, filenames []string, wg *sync.WaitGroup) {
	for _, filename := range filenames {
		asset := theme.Asset{Key: filename}
		wg.Add(1)
		client.DeleteAsset(asset, func(resp *kit.ShopifyResponse, err kit.Error) {
			if err != nil {
				kit.Errorf(err.Error())
			} else {
				kit.Logf(
					"Successfully removed file %s from %s",
					kit.BlueText(resp.Asset.Key),
					kit.YellowText(resp.Host),
				)
				removeFile(resp.Asset.Key)
			}
			wg.Done()
		})
	}
	wg.Done()
}

func removeFile(filename string) error {
	dir, err := os.Getwd()
	err = os.Remove(fmt.Sprintf("%s/%s", dir, filename))
	return err
}
