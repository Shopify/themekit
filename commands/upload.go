package commands

import (
	"github.com/Shopify/themekit"
	"github.com/Shopify/themekit/theme"
)

// UploadCommand add file(s) to theme
func UploadCommand(args Args) chan bool {
	files := make(chan themekit.AssetEvent)
	go ReadAndPrepareFiles(args, files)

	done, events := args.ThemeClient.Process(files)
	mergeEvents(args.EventLog, []chan themekit.ThemeEvent{events})
	return done
}

// ReadAndPrepareFiles ... TODO
func ReadAndPrepareFiles(args Args, results chan themekit.AssetEvent) {
	for _, filename := range args.Filenames {
		asset, err := loadAsset(args, filename)

		if err == nil {
			results <- themekit.NewUploadEvent(asset)
		} else if err.Error() != "File is a directory" {
			themekit.NotifyError(err)
		}
	}
	close(results)
}

func loadAsset(args Args, filename string) (asset theme.Asset, err error) {
	root, err := args.WorkingDirGetter()
	if err != nil {
		return
	}

	return theme.LoadAsset(root, filename)
}
