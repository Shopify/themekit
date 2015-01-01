package main

import (
	"fmt"
	"github.com/csaunders/phoenix"
	"log"
	"os"
)

const DeprecationNotice string = "theme-upload is deprecated. Use `theme-maniuplate upload file(s)` instead"

func main() {
	fmt.Println(phoenix.RedText(DeprecationNotice))
	dir, _ := os.Getwd()
	config, _ := phoenix.LoadConfigurationFromFile(fmt.Sprintf("%s/config.yml", dir))
	files := make(chan phoenix.AssetEvent)
	go readAndPrepareFiles(dir, grabFilenames(), files)

	client := phoenix.NewThemeClient(config)
	done, messages := client.Process(files)
	go func() {
		for {
			message := <-messages
			fmt.Println(message)
		}
	}()
	<-done
}

func grabFilenames() []string {
	names := os.Args[1:]
	if len(names) <= 0 {
		fmt.Println("Missing name of file to upload")
		fmt.Println("Usage: theme-upload filename [filename...]")
		os.Exit(0)
	}
	return names
}

func readAndPrepareFiles(root string, filenames []string, results chan phoenix.AssetEvent) {
	for _, filename := range filenames {
		assetEvent, err := loadAsset(root, filename)
		if err == nil {
			results <- assetEvent
		} else if err.Error() != "File is a directory" {
			log.Panic(err)
		}
	}
	close(results)
}

func loadAsset(root, filename string) (assetEvent phoenix.AssetEvent, err error) {
	asset, err := phoenix.LoadAsset(root, filename)

	if err != nil {
		return
	}

	assetEvent = phoenix.NewUploadEvent(asset)
	return
}
