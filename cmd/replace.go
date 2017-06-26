package cmd

import (
	"fmt"
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
	RunE:     arbiter.forEachClient(replace),
	PostRunE: arbiter.forEachClient(uploadSettingsData),
}

func replace(client kit.ThemeClient, filenames []string) error {
	if len(filenames) > 0 {
		return upload(client, filenames)
	}

	if client.Config.ReadOnly {
		return fmt.Errorf("[%s] environment is reaonly", kit.GreenText(client.Config.Environment))
	}

	assetsActions := map[string]assetAction{}
	assets, remoteErr := client.AssetList()
	if remoteErr != nil {
		return remoteErr
	}

	for _, asset := range assets {
		assetsActions[asset.Key] = assetAction{asset: asset, event: kit.Remove}
	}

	localAssets, localErr := client.LocalAssets()
	if localErr != nil {
		return localErr
	}

	for _, asset := range localAssets {
		assetsActions[asset.Key] = assetAction{asset: asset, event: kit.Update}
	}

	var wg sync.WaitGroup
	bar := arbiter.newProgressBar(len(assetsActions), client.Config.Environment)
	for key, action := range assetsActions {
		if key == settingsDataKey {
			arbiter.incBar(bar) //pretend we did this one we will do it later
			continue
		}
		wg.Add(1)
		go perform(client, action.asset, action.event, bar, &wg)
	}
	wg.Wait()
	return nil
}
