package commands

import (
	"os"

	"github.com/Shopify/themekit/kit"
	"github.com/Shopify/themekit/theme"
)

// UploadCommand add file(s) to theme
func UploadCommand(args Args, done chan bool) {
	foreman := args.ThemeClient.NewForeman()
	logs := args.ThemeClient.Process(foreman.WorkerQueue, done)
	mergeEvents(args.EventLog, []chan kit.ThemeEvent{logs})
	go enqueueUploadEvents(args.ThemeClient, args.Filenames, foreman.JobQueue)
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
