package commands

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Shopify/themekit/kit"
	"github.com/Shopify/themekit/theme"
)

// DownloadCommand downloads file(s) from theme
func DownloadCommand(args Args, done chan bool) {
	if len(args.Filenames) <= 0 {
		go downloadAllFiles(args.ThemeClient, done)
	} else {
		go downloadFiles(args.ThemeClient, args.Filenames, done)
	}
}

func downloadAllFiles(client kit.ThemeClient, done chan bool) {
	assets, errs := client.AssetList()
	go drainErrors(errs)
	for {
		asset, more := <-assets
		if more {
			writeToDisk(client, asset)
		} else {
			done <- true
			return
		}
	}
}

func downloadFiles(client kit.ThemeClient, filenames []string, done chan bool) {
	for _, filename := range filenames {
		if asset, err := client.Asset(filename); err != nil {
			if nonFatal, ok := err.(kit.NonFatalNetworkError); ok {
				client.Message("[%s] Could not complete %s for %s", kit.RedText(fmt.Sprintf("%d", nonFatal.Code)), kit.YellowText(nonFatal.Verb), kit.BlueText(filename))
			}
		} else {
			writeToDisk(client, asset)
		}
	}
	done <- true
	return
}

func writeToDisk(client kit.ThemeClient, asset theme.Asset) {
	dir, err := os.Getwd()
	if err != nil {
		kit.NotifyError(err)
		return
	}

	perms, err := os.Stat(dir)
	if err != nil {
		kit.NotifyError(err)
		return
	}

	filename := fmt.Sprintf("%s/%s", dir, asset.Key)
	err = os.MkdirAll(filepath.Dir(filename), perms.Mode())
	if err != nil {
		kit.NotifyError(err)
		return
	}

	file, err := os.Create(filename)
	if err != nil {
		kit.NotifyError(err)
		return
	}
	defer file.Sync()
	defer file.Close()

	var data []byte
	switch {
	case len(asset.Value) > 0:
		data = []byte(asset.Value)
	case len(asset.Attachment) > 0:
		data, err = base64.StdEncoding.DecodeString(asset.Attachment)
		if err != nil {
			kit.NotifyError(fmt.Errorf("Could not decode %s. error: %s", asset.Key, err))
			return
		}
	}

	if len(data) > 0 {
		_, err = prettyWrite(file, data)
	}

	if err != nil {
		kit.NotifyError(err)
	} else {
		client.Message(kit.GreenText(fmt.Sprintf("Successfully wrote %s to disk", filename)))
	}
}

func prettyWrite(file *os.File, data []byte) (n int, err error) {
	switch filepath.Ext(file.Name()) {
	case ".json":
		var out bytes.Buffer
		json.Indent(&out, data, "", "\t")
		return file.Write(out.Bytes())
	default:
		return file.Write(data)
	}
}

func handleError(filename string, err error) {
}
