package commands

import (
	"os"

	"github.com/Shopify/themekit/kit"
	"github.com/Shopify/themekit/theme"
)

// ReplaceCommand overwrite theme file(s)
func ReplaceCommand(args Args, done chan bool) {
	foreman := args.ThemeClient.NewForeman()
	args.ThemeClient.Process(foreman.WorkerQueue, done)
	root, _ := os.Getwd()
	assetsActions := map[string]kit.AssetEvent{}
	if len(args.Filenames) == 0 {
		for _, asset := range args.ThemeClient.AssetListSync() {
			assetsActions[asset.Key] = kit.NewRemovalEvent(asset)
		}
		for _, asset := range args.ThemeClient.LocalAssets(root) {
			assetsActions[asset.Key] = kit.NewUploadEvent(asset)
		}
	} else {
		for _, filename := range args.Filenames {
			asset, err := theme.LoadAsset(root, filename)
			if err == nil && asset.IsValid() {
				assetsActions[asset.Key] = kit.NewUploadEvent(asset)
			}
		}
	}
	for _, event := range assetsActions {
		foreman.JobQueue <- event
	}
	close(foreman.JobQueue)
}
