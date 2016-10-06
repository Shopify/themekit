package cmd

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
		for _, asset := range args.ThemeClient.AssetList() {
			if err := writeToDisk(args.ThemeClient, asset); err != nil {
				kit.Fatal(err)
			}
		}
	} else {
		for _, filename := range args.Filenames {
			if asset, err := args.ThemeClient.Asset(filename); err != nil {
				if nonFatal, ok := err.(kit.NonFatalNetworkError); ok {
					args.ThemeClient.Message("[%s] Could not complete %s for %s", kit.RedText(fmt.Sprintf("%d", nonFatal.Code)), kit.YellowText(nonFatal.Verb), kit.BlueText(filename))
				} else {
					kit.Fatal(err)
				}
			} else {
				if err := writeToDisk(args.ThemeClient, asset); err != nil {
					kit.Fatal(err)
				}
			}
		}
	}
	done <- true
}

func writeToDisk(client kit.ThemeClient, asset theme.Asset) error {
	dir, err := os.Getwd()
	if err != nil {
		return err
	}

	perms, err := os.Stat(dir)
	if err != nil {
		return err
	}

	filename := fmt.Sprintf("%s/%s", dir, asset.Key)
	err = os.MkdirAll(filepath.Dir(filename), perms.Mode())
	if err != nil {
		return err
	}

	file, err := os.Create(filename)
	if err != nil {
		return err
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
			return fmt.Errorf("Could not decode %s. error: %s", asset.Key, err)
		}
	}

	if len(data) > 0 {
		_, err = prettyWrite(file, data)
	}

	if err != nil {
		return err
	}

	client.Message(kit.GreenText(fmt.Sprintf("Successfully wrote %s to disk", filename)))

	return nil
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
