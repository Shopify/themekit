package cmd

import (
	"os"
	"path/filepath"
	"sync"

	"github.com/spf13/cobra"

	"github.com/Shopify/themekit/kit"
)

var removeCmd = &cobra.Command{
	Use:   "remove <filenames>",
	Short: "Remove theme file(s) from shopify",
	Long:  `Remove will delete all specified files from shopify servers.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		themeClients, err := generateThemeClients()
		if err != nil {
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
	defer wg.Done()
	for _, filename := range filenames {
		wg.Add(1)
		go performRemove(client, kit.Asset{Key: filename}, wg)
	}
}

func performRemove(client kit.ThemeClient, asset kit.Asset, wg *sync.WaitGroup) {
	defer wg.Done()
	resp, err := client.DeleteAsset(asset)
	if err != nil {
		kit.LogError(err)
	} else {
		kit.Printf(
			"Successfully removed file %s from %s",
			kit.BlueText(asset.Key),
			kit.YellowText(resp.Host),
		)
		os.Remove(filepath.Join(directory, asset.Key))
	}
}
