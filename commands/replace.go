package commands

import (
	"os"

	"github.com/Shopify/themekit/kit"
	"github.com/Shopify/themekit/theme"
)

// ReplaceCommand overwrite theme file(s)
func ReplaceCommand(args Args, done chan bool) {
	rawEvents, throttledEvents := prepareChannel(args)
	logs := args.ThemeClient.Process(throttledEvents, done)
	mergeEvents(args.EventLog, []chan kit.ThemeEvent{logs})
	enqueueEvents(args.ThemeClient, args.Filenames, rawEvents)
}

func enqueueEvents(client kit.ThemeClient, filenames []string, events chan kit.AssetEvent) {
	root, _ := os.Getwd()
	if len(filenames) == 0 {
		go fullReplace(client.AssetListSync(), client.LocalAssets(root), events)
		return
	}
	go func() {
		for _, filename := range filenames {
			asset, err := theme.LoadAsset(root, filename)
			if err == nil {
				events <- kit.NewRemovalEvent(asset)
				events <- kit.NewUploadEvent(asset)
			}
		}
		close(events)
	}()
}

// fullReplace takes slices with assets both from the local filesystem and the remote server and translates them
// into a suitable set of events that updates the remote site to the local state.
func fullReplace(remoteAssets, localAssets []theme.Asset, events chan kit.AssetEvent) {
	assetsActions := map[string]kit.AssetEvent{}
	generateActions := func(assets []theme.Asset, assetEventFn func(asset theme.Asset) kit.SimpleAssetEvent) {
		for _, asset := range assets {
			assetsActions[asset.Key] = assetEventFn(asset)
		}
	}
	generateActions(remoteAssets, kit.NewRemovalEvent)
	generateActions(localAssets, kit.NewUploadEvent)
	go func() {
		for _, event := range assetsActions {
			events <- event
		}
		close(events)
	}()

}

func prepareChannel(args Args) (rawEvents, throttledEvents chan kit.AssetEvent) {
	rawEvents = make(chan kit.AssetEvent)
	if args.Bucket == nil {
		return rawEvents, rawEvents
	}

	foreman := kit.NewForeman(args.Bucket)
	foreman.JobQueue = rawEvents
	foreman.WorkerQueue = make(chan kit.AssetEvent)
	foreman.IssueWork()
	return foreman.JobQueue, foreman.WorkerQueue
}
