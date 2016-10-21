package cmd

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/Shopify/themekit/kit"
)

var downloadCmd = &cobra.Command{
	Use:   "download <filenames>",
	Short: "Download one or all of the theme files",
	Long: `Download will download specific files from shopify servers if provided file names.
If no filenames are provided then download will download every file in the project
and write them to disk.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := initializeConfig(cmd.Name(), true); err != nil {
			return err
		}
		return download(themeClients[0], args)
	},
}

func download(client kit.ThemeClient, filenames []string) error {
	if len(filenames) <= 0 {
		kit.Printf("Fetching assets from %s", kit.YellowText(client.Config.Domain))
		assets, err := client.AssetList()
		if err != nil {
			return err
		}
		for _, asset := range assets {
			if err := writeToDisk(client, asset); err != nil {
				return err
			}
		}
	} else {
		for _, filename := range filenames {
			kit.Printf("Fetching %s from %s", filename, kit.YellowText(client.Config.Domain))
			asset, err := client.Asset(filename)
			if err != nil {
				return err
			}
			if err := writeToDisk(client, asset); err != nil {
				return err
			}
		}
	}
	return nil
}

func writeToDisk(client kit.ThemeClient, asset kit.Asset) error {
	dir := client.Config.Directory
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
	kit.LogNotifyf("Successfully wrote %s to disk", filename)
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