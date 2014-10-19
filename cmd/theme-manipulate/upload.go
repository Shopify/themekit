package main

import (
	"fmt"
	"github.com/csaunders/phoenix"
	"log"
)

func UploadOperation(client phoenix.ThemeClient, filenames []string) (done chan bool) {
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
		if err != nil {
			log.Panic(err)
		}
		results <- phoenix.NewUploadEvent(asset)
	}
	close(results)
}
