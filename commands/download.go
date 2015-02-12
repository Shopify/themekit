package commands

import (
	"encoding/base64"
	"fmt"
	"github.com/csaunders/phoenix"
	"log"
	"os"
	"path/filepath"
)

func DownloadOperation(client phoenix.ThemeClient, filenames []string) (done chan bool) {
	done = make(chan bool)

	if len(filenames) <= 0 {
		go downloadAllFiles(client.AssetList(), done)
	} else {
		go downloadFiles(client.Asset, filenames, done)
	}

	return done
}

func downloadAllFiles(assets chan phoenix.Asset, done chan bool) {
	for {
		asset, more := <-assets
		if more {
			writeToDisk(asset)
		} else {
			done <- true
			return
		}
	}
}

func downloadFiles(retrievalFunction phoenix.AssetRetrieval, filenames []string, done chan bool) {
	for _, filename := range filenames {
		asset := retrievalFunction(filename)
		writeToDisk(asset)
	}
	done <- true
	return
}

func writeToDisk(asset phoenix.Asset) {
	dir, err := os.Getwd()
	if err != nil {
		log.Fatal("Could not get current working directory ", err)
	}

	perms, err := os.Stat(dir)
	if err != nil {
		log.Fatal("Could not get directory information ", err)
	}

	filename := fmt.Sprintf("%s/%s", dir, asset.Key)
	err = os.MkdirAll(filepath.Dir(filename), perms.Mode())
	if err != nil {
		log.Fatal("Could not create parent directory ", err)
	}

	file, err := os.Create(filename)
	defer file.Sync()
	defer file.Close()
	if err != nil {
		log.Fatal("Could not create ", filename, err)
	}

	var data []byte
	switch {
	case len(asset.Value) > 0:
		data = []byte(asset.Value)
	case len(asset.Attachment) > 0:
		data, err = base64.StdEncoding.DecodeString(asset.Attachment)
		if err != nil {
			fmt.Println(fmt.Sprintf("Could not decode %s. error: %s", asset.Key, err))
			return
		}
	}

	if len(data) > 0 {
		_, err = file.Write(data)
	}

	if err != nil {
		log.Fatal("Could not write file to disk ", err)
	}
}
