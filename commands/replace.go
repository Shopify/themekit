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
	enqueueReplaceEvents(args.ThemeClient, args.Filenames, foreman.JobQueue)
}

func enqueueReplaceEvents(client kit.ThemeClient, filenames []string, events chan kit.AssetEvent) {
	root, _ := os.Getwd()
	assetsActions := map[string]kit.AssetEvent{}
	if len(filenames) == 0 {
		for _, asset := range client.AssetListSync() {
			assetsActions[asset.Key] = kit.NewRemovalEvent(asset)
		}
		for _, asset := range client.LocalAssets(root) {
			assetsActions[asset.Key] = kit.NewUploadEvent(asset)
		}
	} else {
		for _, filename := range filenames {
			asset, err := theme.LoadAsset(root, filename)
			if err == nil && asset.IsValid() {
				assetsActions[asset.Key] = kit.NewUploadEvent(asset)
			}
		}
	}
	for _, event := range assetsActions {
		events <- event
	}
	close(events)
}
