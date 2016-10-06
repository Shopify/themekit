package cmd

import (
	"os"

	"github.com/Shopify/themekit/kit"
	"github.com/Shopify/themekit/theme"
)

// UploadCommand add file(s) to theme
func UploadCommand(args Args, done chan bool) {
	jobQueue := args.ThemeClient.Process(done)
	root, _ := os.Getwd()
	if len(args.Filenames) == 0 {
		for _, asset := range args.ThemeClient.LocalAssets(root) {
			if asset.IsValid() {
				jobQueue <- kit.NewUploadEvent(asset)
			}
		}
	} else {
		for _, filename := range args.Filenames {
			asset, err := theme.LoadAsset(root, filename)
			if err != nil {
				args.ThemeClient.ErrorMessage(err.Error())
			} else if asset.IsValid() {
				jobQueue <- kit.NewUploadEvent(asset)
			}
		}
	}
	close(jobQueue)
}
