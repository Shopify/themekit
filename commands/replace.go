package commands

import (
	"github.com/Shopify/themekit"
	"github.com/Shopify/themekit/bucket"
	"github.com/Shopify/themekit/theme"
	"os"
)

type ReplaceOptions struct {
	BasicOptions
	Bucket *bucket.LeakyBucket
}

func ReplaceCommand(args map[string]interface{}) chan bool {
	options := ReplaceOptions{}
	extractThemeClient(&options.Client, args)
	extractEventLog(&options.EventLog, args)
	options.Filenames = extractStringSlice("filenames", args)

	return Replace(options)
}

func Replace(options ReplaceOptions) chan bool {
	rawEvents, throttledEvents := prepareChannel(options)
	done, logs := options.Client.Process(throttledEvents)
	mergeEvents(options.getEventLog(), []chan themekit.ThemeEvent{logs})
	enqueueEvents(options.Client, options.Filenames, rawEvents)
	return done
}

func enqueueEvents(client themekit.ThemeClient, filenames []string, events chan themekit.AssetEvent) {
	root, _ := os.Getwd()
	if len(filenames) == 0 {
		go fullReplace(client.AssetListSync(), client.LocalAssets(root), events)
		return
	}
	go func() {
		for _, filename := range filenames {
			asset, err := theme.LoadAsset(root, filename)
			if err == nil {
				events <- themekit.NewUploadEvent(asset)
			}
		}
		close(events)
	}()
}

// fullReplace takes slices with assets both from the local filesystem and the remote server and translates them
// into a suitable set of events that updates the remote site to the local state.
func fullReplace(remoteAssets, localAssets []theme.Asset, events chan themekit.AssetEvent) {
	assetsActions := map[string]themekit.AssetEvent{}
	generateActions := func(assets []theme.Asset, assetEventFn func(asset theme.Asset) themekit.SimpleAssetEvent) {
		for _, asset := range assets {
			assetsActions[asset.Key] = assetEventFn(asset)
		}
	}
	generateActions(remoteAssets, themekit.NewRemovalEvent)
	generateActions(localAssets, themekit.NewUploadEvent)
	go func() {
		for _, event := range assetsActions {
			events <- event
		}
		close(events)
	}()

}

func prepareChannel(options ReplaceOptions) (rawEvents, throttledEvents chan themekit.AssetEvent) {
	rawEvents = make(chan themekit.AssetEvent)
	if options.Bucket == nil {
		return rawEvents, rawEvents
	}

	foreman := themekit.NewForeman(options.Bucket)
	foreman.JobQueue = rawEvents
	foreman.WorkerQueue = make(chan themekit.AssetEvent)
	foreman.IssueWork()
	return foreman.JobQueue, foreman.WorkerQueue
}
