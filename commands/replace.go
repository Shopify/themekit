package commands

import (
	"fmt"
	"github.com/csaunders/phoenix"
)

func ReplaceCommand(args map[string]interface{}) chan bool {
	return toClientAndFilesAsync(args, Replace)
}

func Replace(client phoenix.ThemeClient, filenames []string) chan bool {
	events := make(chan phoenix.AssetEvent)
	done, messages := client.Process(events)

	go func() {
		for {
			fmt.Println(<-messages)
		}
	}()

	if assets, err := assetList(client, filenames); err != nil {
		phoenix.NotifyError(err)
	} else {
		go removeAndUpload(assets, events)
	}
	return done
}

func assetList(client phoenix.ThemeClient, filenames []string) (chan phoenix.Asset, error) {
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
	return assets, nil
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
