package cmd

import (
	"sync"

	"github.com/spf13/cobra"

	"github.com/Shopify/themekit/kit"
)

var replaceCmd = &cobra.Command{
	Use:   "replace <filenames>",
	Short: "Overwrite theme file(s)",
	Long: `Replace will overwrite specific files if provided with file names.
If replace is not provided with file names then it will replace all
the files on shopify with your local files. Any files that do not
exist on your local machine will be removed from shopify.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := initializeConfig(cmd.Name(), true); err != nil {
			return err
		}

		wg := sync.WaitGroup{}
		for _, client := range themeClients {
			wg.Add(1)
			go replace(client, args, &wg)
		}
		wg.Wait()
		return nil
	},
}

func replace(client kit.ThemeClient, filenames []string, wg *sync.WaitGroup) error {
	assetsActions := map[string]kit.AssetEvent{}
	if len(filenames) == 0 {
		assets, remoteErr := client.AssetList()
		if remoteErr != nil {
			return remoteErr
		}

		for _, asset := range assets {
			assetsActions[asset.Key] = kit.AssetEvent{Asset: asset, Type: kit.Remove}
		}

		localAssets, localErr := client.LocalAssets()
		if localErr != nil {
			return localErr
		}

		for _, asset := range localAssets {
			assetsActions[asset.Key] = kit.AssetEvent{Asset: asset, Type: kit.Update}
		}
	} else {
		for _, filename := range filenames {
			asset, err := client.LocalAsset(filename)
			if err != nil {
				return err
			} else {
				assetsActions[asset.Key] = kit.AssetEvent{Asset: asset, Type: kit.Update}
			}
		}
	}
	for _, event := range assetsActions {
		performReplace(client, event, wg)
	}
	wg.Done()
	return nil
}

func performReplace(client kit.ThemeClient, event kit.AssetEvent, wg *sync.WaitGroup) {
	wg.Add(1)
	client.Perform(event, func(resp *kit.ShopifyResponse, err kit.Error) {
		if err != nil {
			kit.Errorf(err.Error())
		} else {
			kit.Logf(
				"Successfully performed %s on file %s from %s",
				kit.GreenText(resp.EventType),
				kit.GreenText(resp.Asset.Key),
				kit.YellowText(resp.Host),
			)
		}
		wg.Done()
	})
}
