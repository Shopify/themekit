package commands

import (
	"github.com/csaunders/phoenix"
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
	files := make(chan phoenix.AssetEvent)
	go readAndPrepareFiles(options.Filenames, files)

	done, events := options.Client.Process(files)
	mergeEvents(options.getEventLog(), []chan phoenix.ThemeEvent{events})
	return done
}

func readAndPrepareFiles(filenames []string, results chan phoenix.AssetEvent) {
	for _, filename := range filenames {
		asset, err := loadAsset(filename)
		if err == nil {
			results <- phoenix.NewUploadEvent(asset)
		} else if err.Error() != "File is a directory" {
			phoenix.NotifyError(err)
		}
	}
	close(results)
}

func loadAsset(filename string) (asset phoenix.Asset, err error) {
	root, err := os.Getwd()
	if err != nil {
		return
	}

	return phoenix.LoadAsset(root, filename)
}
