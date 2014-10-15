package main

import (
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
		log.Fatal(err)
	}

	perms, err := os.Stat(dir)
	if err != nil {
		log.Fatal(err)
	}

	filename := dir + asset.Key
	err = os.MkdirAll(filepath.Dir(filename), perms.Mode())
	if err != nil {
		log.Fatal(err)
	}

	file, err := os.Create(filename)
	if err != nil {
		log.Fatal(err)
	}
	_, err = file.Write([]byte(asset.Value))

	if err != nil {
		log.Fatal(err)
	}
}
