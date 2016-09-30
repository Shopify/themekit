package commands

import (
	"os"

	"github.com/Shopify/themekit/kit"
	"github.com/Shopify/themekit/theme"
)

// UploadCommand add file(s) to theme
func UploadCommand(args Args, done chan bool) {
	foreman := args.ThemeClient.NewForeman()
	args.ThemeClient.Process(foreman.WorkerQueue, done)
	root, _ := os.Getwd()
	if len(args.Filenames) == 0 {
		for _, asset := range args.ThemeClient.LocalAssets(root) {
			if asset.IsValid() {
				foreman.JobQueue <- kit.NewUploadEvent(asset)
			}
		}
	} else {
		for _, filename := range args.Filenames {
			asset, err := theme.LoadAsset(root, filename)
			if err == nil && asset.IsValid() {
				foreman.JobQueue <- kit.NewUploadEvent(asset)
			}
		}
	}
	close(foreman.JobQueue)
}
