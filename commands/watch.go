package commands

import (
	"os"

	"github.com/Shopify/themekit/kit"
	"github.com/Shopify/themekit/theme"
)

// WatchCommand watches directories for changes, and updates the remote theme
func WatchCommand(args Args, done chan bool) {
	if len(args.ThemeClients) == 0 {
		args.ThemeClients = []kit.ThemeClient{args.ThemeClient}
	}

	eventLog := args.EventLog

	for _, client := range args.ThemeClients {
		config := client.GetConfiguration()
		concurrency := config.Concurrency
		logEvent(message(fmt.Sprintf("Spawning %d workers for %s", concurrency, config.Domain)), eventLog)

		args.ThemeClient = client
		watchForChangesAndIssueWork(args, eventLog)
	}
}

// UploadCommand add file(s) to theme
func UploadCommand(args Args, done chan bool) {
	rawEvents, throttledEvents := prepareChannel(args)
	logs := args.ThemeClient.Process(throttledEvents, done)
	mergeEvents(args.EventLog, []chan kit.ThemeEvent{logs})
	go enqueueUploadEvents(args.ThemeClient, args.Filenames, rawEvents)
}

func enqueueUploadEvents(client kit.ThemeClient, filenames []string, events chan kit.AssetEvent) {
	root, _ := os.Getwd()
	if len(filenames) == 0 {
		for _, asset := range client.LocalAssets(root) {
			events <- kit.NewUploadEvent(asset)
		}
	} else {
		for _, filename := range filenames {
			asset, err := theme.LoadAsset(root, filename)
			if err == nil {
				events <- kit.NewUploadEvent(asset)
			}
		}
	}
	close(events)
}
