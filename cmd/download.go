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
		themeClients, err := generateThemeClients()
		if err != nil {
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
			if err := writeToDisk(client.Config.Directory, asset); err != nil {
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
			if err := writeToDisk(client.Config.Directory, asset); err != nil {
				return err
			}
		}
	}
	return nil
}

func writeToDisk(directory string, asset kit.Asset) error {
	perms, err := os.Stat(directory)
	if err != nil {
		return err
	}

	filename := filepath.Join(directory, asset.Key)
	err = os.MkdirAll(filepath.Dir(filename), perms.Mode())
	if err != nil {
		return err
	}

	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	contents, err := getAssetContents(asset)
	if err != nil {
		return err
	}

	if _, err = formatWrite(file, contents); err != nil {
		return err
	}

	kit.LogNotifyf("Successfully wrote %s to disk", filename)
	return nil
}

func getAssetContents(asset kit.Asset) ([]byte, error) {
	var data []byte
	var err error
	switch {
	case len(asset.Value) > 0:
		data = []byte(asset.Value)
	case len(asset.Attachment) > 0:
		if data, err = base64.StdEncoding.DecodeString(asset.Attachment); err != nil {
			return data, fmt.Errorf("Could not decode %s. error: %s", asset.Key, err)
		}
	}
	return data, nil
}

func formatWrite(file *os.File, data []byte) (n int, err error) {
	if len(data) == 0 {
		return 0, nil
	}

	defer file.Sync()
	switch filepath.Ext(file.Name()) {
	case ".json":
		var out bytes.Buffer
		json.Indent(&out, data, "", "\t")
		return file.Write(out.Bytes())
	default:
		return file.Write(data)
	}
}
