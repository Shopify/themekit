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
		message(args.EventLog, "Spawning %d workers for %s", config.Concurrency, config.Domain)
		foreman := buildForeman(client, args, config)
		for i := 0; i < config.Concurrency; i++ {
			go spawnWorker(foreman.WorkerQueue, client, args.EventLog)
			message(args.EventLog, "%s Worker #%d ready to upload local changes", config.Domain, i)
		}
	}
}

func buildForeman(client kit.ThemeClient, args Args, config kit.Configuration) kit.Foreman {
	foreman := client.NewForeman()
	if len(args.NotifyFile) > 0 {
		foreman.OnIdle = func() {
			os.Create(args.NotifyFile)
			os.Chtimes(args.NotifyFile, time.Now(), time.Now())
		}
	}
	foreman.JobQueue = constructFileWatcher(args.Directory, config)
	foreman.Restart()
	return foreman
}

func constructFileWatcher(dir string, config kit.Configuration) chan kit.AssetEvent {
	filter := kit.NewEventFilterFromPatternsAndFiles(config.IgnoredFiles, config.Ignores)
	watcher, err := kit.NewFileWatcher(dir, true, filter)
	if err != nil {
		kit.NotifyError(err)
	}
	return watcher
}

func spawnWorker(queue chan kit.AssetEvent, client kit.ThemeClient, eventLog chan kit.ThemeEvent) {
	for {
		asset := <-queue
		if asset.Asset().IsValid() {
			message(eventLog, "Received %s event on %s", kit.GreenText(asset.Type().String()), kit.BlueText(asset.Asset().Key))
			logEvent(client.Perform(asset), eventLog)
		}
	}
}
