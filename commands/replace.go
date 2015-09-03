package commands

import (
	"github.com/Shopify/themekit"
	"os"
)

type ReplaceOptions struct {
	BasicOptions
	Bucket *themekit.LeakyBucket
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

	assets, errs := assetList(options.Client, options.Filenames)
	go drainErrors(errs)
	go removeAndUpload(assets, rawEvents)

	return done
}

func assetList(client themekit.ThemeClient, filenames []string) (chan themekit.Asset, chan error) {
	if len(filenames) == 0 {
		return client.AssetList()
	}

	assets := make(chan themekit.Asset)
	errs := make(chan error)
	close(errs)
	go func() {
		root, _ := os.Getwd()
		for _, filename := range filenames {
			asset, _ := themekit.LoadAsset(root, filename)
			assets <- asset
		}
		close(assets)
	}()
	return assets, errs
}

func removeAndUpload(assets chan themekit.Asset, assetEvents chan themekit.AssetEvent) {
	for {
		asset, more := <-assets
		if more {
			assetEvents <- themekit.NewRemovalEvent(asset)
			assetEvents <- themekit.NewUploadEvent(asset)
		} else {
			close(assetEvents)
			return
		}
	}
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
