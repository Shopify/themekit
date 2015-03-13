package commands

import (
	"github.com/csaunders/phoenix"
	"os"
)

func UploadCommand(args map[string]interface{}) chan bool {
	return toClientAndFilesAsync(args, Upload)
}

func Upload(client phoenix.ThemeClient, filenames []string) chan bool {
	files := make(chan phoenix.AssetEvent)
	go readAndPrepareFiles(filenames, files)

	done, _ := client.Process(files)
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
