package commands

import (
	"github.com/csaunders/phoenix"
)

type ReplaceOptions struct {
	BasicOptions
}

func ReplaceCommand(args map[string]interface{}) chan bool {
	options := ReplaceOptions{}
	extractThemeClient(&options.Client, args)
	extractEventLog(&options.EventLog, args)
	options.Filenames = extractStringSlice("filenames", args)

	return Replace(options)
}

func Replace(options ReplaceOptions) chan bool {
	events := make(chan phoenix.AssetEvent)
	done, logs := options.Client.Process(events)
	mergeEvents(options.getEventLog(), []chan phoenix.ThemeEvent{logs})

	assets, errs := assetList(options.Client, options.Filenames)
	go drainErrors(errs)
	go removeAndUpload(assets, events)

	return done
}

func assetList(client phoenix.ThemeClient, filenames []string) (chan phoenix.Asset, chan error) {
	if len(filenames) == 0 {
		return client.AssetList()
	}

	assets := make(chan phoenix.Asset)
	errs := make(chan error)
	close(errs)
	go func() {
		for _, filename := range filenames {
			asset := phoenix.Asset{Key: filename}
			assets <- asset
		}
		close(assets)
	}()
	return assets, errs
}

func removeAndUpload(assets chan phoenix.Asset, assetEvents chan phoenix.AssetEvent) {
	for {
		asset, more := <-assets
		if more {
			assetEvents <- phoenix.NewRemovalEvent(asset)
			assetEvents <- phoenix.NewUploadEvent(asset)
		} else {
			close(assetEvents)
			return
		}
	}
}
