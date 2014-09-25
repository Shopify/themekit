package main

import (
	"fmt"
	"github.com/csaunders/phoenix"
	"log"
	"os"
)

func main() {
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
		} else {
			log.Panic(err)
		}
	}
	close(results)
}

func loadAsset(root, filename string) (assetEvent phoenix.AssetEvent, err error) {
	path := fmt.Sprintf("%s/%s", root, filename)
	file, err := os.Open(path)
	info, err := os.Stat(path)
	if err != nil {
		return
	}

	buffer := make([]byte, info.Size())
	_, err = file.Read(buffer)
	if err != nil {
		return
	}
	asset := phoenix.Asset{Value: string(buffer), Key: filename}
	assetEvent = phoenix.NewUploadEvent(asset)
	return
}
