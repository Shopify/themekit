package commands

import (
	"os"

	"github.com/Shopify/themekit/kit"
	"github.com/Shopify/themekit/theme"
)

// UploadCommand add file(s) to theme
func UploadCommand(args Args, done chan bool) {
	rawEvents, throttledEvents := prepareChannel(args)
	logs := args.ThemeClient.Process(throttledEvents, done)
	mergeEvents(args.EventLog, []chan kit.ThemeEvent{logs})
	enqueueUploadEvents(args.ThemeClient, args.Filenames, rawEvents)
}

func enqueueUploadEvents(client kit.ThemeClient, filenames []string, events chan kit.AssetEvent) {
	root, _ := os.Getwd()
	if len(filenames) == 0 {
		go fullUpload(client.LocalAssets(root), events)
		return
	}
	go func() {
		for _, filename := range filenames {
			asset, err := theme.LoadAsset(root, filename)
			if err == nil {
				events <- kit.NewUploadEvent(asset)
			}
		}
		close(events)
	}()
}

func fullUpload(localAssets []theme.Asset, events chan kit.AssetEvent) {
	assetsActions := map[string]kit.AssetEvent{}
	generateActions := func(assets []theme.Asset, assetEventFn func(asset theme.Asset) kit.SimpleAssetEvent) {
		for _, asset := range assets {
			assetsActions[asset.Key] = assetEventFn(asset)
		}
	}
	generateActions(localAssets, kit.NewUploadEvent)
	go func() {
		for _, event := range assetsActions {
			events <- event
		}
		close(events)
	}()
}
