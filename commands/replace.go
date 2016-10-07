package commands

import (
	"os"

	"github.com/Shopify/themekit/kit"
	"github.com/Shopify/themekit/theme"
)

// ReplaceCommand overwrite theme file(s)
func ReplaceCommand(args Args, done chan bool) {
	jobQueue := args.ThemeClient.Process(done)
	root, _ := os.Getwd()
	assetsActions := map[string]kit.AssetEvent{}
	if len(args.Filenames) == 0 {
		for _, asset := range args.ThemeClient.AssetList() {
			assetsActions[asset.Key] = kit.NewRemovalEvent(asset)
		}
		for _, asset := range args.ThemeClient.LocalAssets(root) {
			assetsActions[asset.Key] = kit.NewUploadEvent(asset)
		}
	} else {
		for _, filename := range args.Filenames {
			asset, err := theme.LoadAsset(root, filename)
			if err != nil {
				args.ThemeClient.ErrorMessage(err.Error())
			} else if asset.IsValid() {
				assetsActions[asset.Key] = kit.NewUploadEvent(asset)
			}
		}
	}
	for _, event := range assetsActions {
		jobQueue <- event
	}
	close(jobQueue)
}
