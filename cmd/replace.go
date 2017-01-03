package cmd

import (
	"sync"

	"github.com/spf13/cobra"

	"github.com/Shopify/themekit/kit"
)

type assetAction struct {
	asset kit.Asset
	event kit.EventType
}

var replaceCmd = &cobra.Command{
	Use:   "replace <filenames>",
	Short: "Overwrite theme file(s)",
	Long: `Replace will overwrite specific files if provided with file names.
If replace is not provided with file names then it will replace all
the files on shopify with your local files. Any files that do not
exist on your local machine will be removed from shopify.

For more documentation please see http://shopify.github.io/themekit/commands/#replace
`,
	RunE: forEachClient(replace, uploadSettingsData),
}

func replace(client kit.ThemeClient, filenames []string, wg *sync.WaitGroup) {
	defer wg.Done()

	if client.Config.ReadOnly {
		kit.LogErrorf("[%s]environment is reaonly", kit.GreenText(client.Config.Environment))
		return
	}

	assetsActions := map[string]assetAction{}

	if len(filenames) == 0 {
		assets, remoteErr := client.AssetList()
		if remoteErr != nil {
			kit.LogError(remoteErr)
			return
		}

		for _, asset := range assets {
			assetsActions[asset.Key] = assetAction{asset: asset, event: kit.Remove}
		}

		localAssets, localErr := client.LocalAssets()
		if localErr != nil {
			kit.LogError(localErr)
			return
		}

		for _, asset := range localAssets {
			assetsActions[asset.Key] = assetAction{asset: asset, event: kit.Update}
		}
	} else {
		for _, filename := range filenames {
			asset, err := client.LocalAsset(filename)
			if err != nil {
				kit.LogError(err)
				return
			}
			assetsActions[asset.Key] = assetAction{asset: asset, event: kit.Update}
		}
	}

	for key, action := range assetsActions {
		if key == settingsDataKey {
			continue
		}
		wg.Add(1)
		go performReplace(client, action.asset, action.event, wg)
	}
}

func performReplace(client kit.ThemeClient, asset kit.Asset, event kit.EventType, wg *sync.WaitGroup) {
	resp, err := client.Perform(asset, event)
	if err != nil {
		kit.LogErrorf("[%s]%s", kit.GreenText(client.Config.Environment), err)
	} else {
		kit.Printf(
			"[%s] Successfully performed %s on file %s from %s",
			kit.GreenText(client.Config.Environment),
			kit.GreenText(resp.EventType),
			kit.GreenText(resp.Asset.Key),
			kit.YellowText(resp.Host),
		)
	}
	wg.Done()
}
