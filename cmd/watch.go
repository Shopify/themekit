package cmd

import (
	"github.com/Shopify/themekit/kit"
)

// WatchCommand watches directories for changes, and updates the remote theme
func WatchCommand(args Args, done chan bool) {
	for _, client := range args.ThemeClients {
		config := client.GetConfiguration()
		client.Message("Spawning %d workers for %s", config.Concurrency, args.Domain)
		assetEvents := client.NewFileWatcher(args.Directory, args.NotifyFile)
		for i := 0; i < config.Concurrency; i++ {
			go spawnWorker(assetEvents, client)
			client.Message("%s Worker #%d ready to upload local changes", config.Domain, i)
		}
	}
}

func spawnWorker(assetEvents chan kit.AssetEvent, client kit.ThemeClient) {
	for {
		event := <-assetEvents
		if event.Asset().IsValid() || event.Type() == kit.Remove {
			client.Message("Received %s event on %s", kit.GreenText(event.Type().String()), kit.BlueText(event.Asset().Key))
			client.Perform(event)
		}
	}
}
