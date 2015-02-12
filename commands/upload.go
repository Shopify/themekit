package commands

import (
	"fmt"
	"github.com/csaunders/phoenix"
	"log"
	"os"
)

func UploadCommand(args map[string]interface{}) chan bool {
	return toClientAndFilesAsync(args, Upload)
}

func Upload(client phoenix.ThemeClient, filenames []string) (done chan bool) {
	files := make(chan phoenix.AssetEvent)
	go readAndPrepareFiles(filenames, files)

	done, messages := client.Process(files)
	go func() {
		for {
			message, more := <-messages
			if !more {
				return
			}
			fmt.Println(message)
		}
	}()

	return
}

func readAndPrepareFiles(filenames []string, results chan phoenix.AssetEvent) {
	for _, filename := range filenames {
		asset, err := loadAsset(filename)
		if err == nil {
			results <- phoenix.NewUploadEvent(asset)
		} else if err.Error() != "File is a directory" {
			log.Panic(err)
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
