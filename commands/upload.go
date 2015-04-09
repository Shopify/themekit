package commands

import (
	"github.com/csaunders/themekit"
	"os"
)

type UploadOptions struct {
	BasicOptions
}

func UploadCommand(args map[string]interface{}) chan bool {
	options := ReplaceOptions{}
	extractThemeClient(&options.Client, args)
	extractEventLog(&options.EventLog, args)
	options.Filenames = extractStringSlice("filenames", args)

	return Upload(options)
}

func Upload(options ReplaceOptions) chan bool {
	files := make(chan themekit.AssetEvent)
	go readAndPrepareFiles(options.Filenames, files)

	done, events := options.Client.Process(files)
	mergeEvents(options.getEventLog(), []chan themekit.ThemeEvent{events})
	return done
}

func readAndPrepareFiles(filenames []string, results chan themekit.AssetEvent) {
	for _, filename := range filenames {
		asset, err := loadAsset(filename)
		if err == nil {
			results <- themekit.NewUploadEvent(asset)
		} else if err.Error() != "File is a directory" {
			themekit.NotifyError(err)
		}
	}
	close(results)
}

func loadAsset(filename string) (asset themekit.Asset, err error) {
	root, err := os.Getwd()
	if err != nil {
		return
	}

	return themekit.LoadAsset(root, filename)
}
