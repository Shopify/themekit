package commands

import (
	"github.com/csaunders/themekit"
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
	events := make(chan themekit.AssetEvent)
	done, logs := options.Client.Process(events)
	mergeEvents(options.getEventLog(), []chan themekit.ThemeEvent{logs})

	assets, errs := assetList(options.Client, options.Filenames)
	go drainErrors(errs)
	go removeAndUpload(assets, events)

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
		for _, filename := range filenames {
			asset := themekit.Asset{Key: filename}
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
