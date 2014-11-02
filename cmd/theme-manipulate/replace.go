package main

import (
	"fmt"
	"github.com/csaunders/phoenix"
)

func ReplaceOperation(client phoenix.ThemeClient, filenames []string) chan bool {
	var assets chan phoenix.Asset
	events := make(chan phoenix.AssetEvent)
	done, messages := client.Process(events)

	go logMessages(messages)

	assets = assetList(client, filenames)
	go removeAndUpload(assets, events)
	return done
}

func assetList(client phoenix.ThemeClient, filenames []string) chan phoenix.Asset {
	if len(filenames) == 0 {
		return client.AssetList()
	}

	assets := make(chan phoenix.Asset)
	go func() {
		for _, filename := range filenames {
			asset := phoenix.Asset{Key: filename}
			assets <- asset
		}
		close(assets)
	}()
	return assets
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

func logMessages(messages chan string) {
	var message string
	for {
		message = <-messages
		if len(message) <= 0 {
			return
		}
		fmt.Println(message)
	}
}
