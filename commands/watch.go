package commands

import (
	"os"
	"time"

	"github.com/Shopify/themekit/kit"
)

// WatchCommand watches directories for changes, and updates the remote theme
func WatchCommand(args Args, done chan bool) {
	for _, client := range args.ThemeClients {
		config := client.GetConfiguration()
		client.Message("Spawning %d workers for %s", config.Concurrency, config.Domain)
		foreman := buildForeman(client, args, config)
		for i := 0; i < config.Concurrency; i++ {
			go spawnWorker(foreman.WorkerQueue, client)
			client.Message("%s Worker #%d ready to upload local changes", config.Domain, i)
		}
	}
}

func buildForeman(client kit.ThemeClient, args Args, config kit.Configuration) *kit.Foreman {
	foreman := client.NewForeman()
	if len(args.NotifyFile) > 0 {
		foreman.OnIdle = func() {
			os.Create(args.NotifyFile)
			os.Chtimes(args.NotifyFile, time.Now(), time.Now())
		}
	}
	var err error
	foreman.JobQueue, err = client.NewFileWatcher(args.Directory)
	if err != nil {
		kit.NotifyError(err)
	}
	foreman.Restart()
	return foreman
}

func spawnWorker(queue chan kit.AssetEvent, client kit.ThemeClient) {
	for {
		asset := <-queue
		if asset.Asset().IsValid() {
			client.Message("Received %s event on %s", kit.GreenText(asset.Type().String()), kit.BlueText(asset.Asset().Key))
			client.Perform(asset)
		}
	}
}
