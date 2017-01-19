package kit

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// Asset represents an asset from the shopify server.
type Asset struct {
	Key         string `json:"key"`
	Value       string `json:"value,omitempty"`
	Attachment  string `json:"attachment,omitempty"`
	ContentType string `json:"content_type,omitempty"`
	ThemeID     int64  `json:"theme_id,omitempty"`
}

// IsValid verifies that the Asset has a Key, and at least a Value or Attachment
func (asset Asset) IsValid() bool {
	return len(asset.Key) > 0 && (len(asset.Value) > 0 || len(asset.Attachment) > 0)
}

// Size will return the length of the value or attachment depending on which exists.
func (asset Asset) Size() int {
	if len(asset.Value) > 0 {
		return len(asset.Value)
	}
	return len(asset.Attachment)
}

// Write will write the asset out to the destination directory
func (asset Asset) Write(directory string) error {
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
	defer file.Sync()
	defer file.Close()

	contents, err := asset.Contents()
	if err != nil {
		return err
	}

	_, err = file.Write(contents)
	return err
}

// Contents will return a byte array of data for the asset contents
func (asset Asset) Contents() ([]byte, error) {
	var data []byte
	var err error
	switch {
	case len(asset.Value) > 0:
		data = []byte(asset.Value)
		if filepath.Ext(asset.Key) == ".json" {
			var out bytes.Buffer
			json.Indent(&out, data, "", "  ")
			data = out.Bytes()
		}
	case len(asset.Attachment) > 0:
		if data, err = base64.StdEncoding.DecodeString(asset.Attachment); err != nil {
			return data, fmt.Errorf("Could not decode %s. error: %s", asset.Key, err)
		}
	}
	return data, nil
}

// ByAsset implements sort.Interface for sorting remote assets
type ByAsset []Asset

// Len returns the length of the array.
func (assets ByAsset) Len() int {
	return len(assets)
}

// Swap swaps two values in the slive
func (assets ByAsset) Swap(i, j int) {
	assets[i], assets[j] = assets[j], assets[i]
}

// Less is the comparison method. Will return true if the first is less than the second.
func (assets ByAsset) Less(i, j int) bool {
	return assets[i].Key < assets[j].Key
}

func findAllFiles(dir string) ([]string, error) {
	var files []string

	if fileInfo, err := os.Stat(dir); err != nil || !fileInfo.IsDir() {
		return files, fmt.Errorf("Path is not a directory")
	}

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}

		files = append(files, path)
		return nil
	})

	return files, err
}

func loadAssetsFromDirectory(dir string, ignore func(path string) bool) ([]Asset, error) {
	assets := []Asset{}
	files, err := findAllFiles(dir)
	if err != nil {
		return assets, err
	}

	for _, file := range files {
		assetKey, err := filepath.Rel(dir, file)
		if err != nil {
			return assets, err
		}
		if !ignore(assetKey) {
			asset, err := loadAsset(dir, assetKey)
			if err == nil {
				assets = append(assets, asset)
			}
		}
	}

	return assets, nil
}

func loadAsset(root, filename string) (asset Asset, err error) {
	asset = Asset{}
	path := filepath.ToSlash(filepath.Join(root, filename))
	path = strings.Replace(path, "\\", "/", -1)
	file, err := os.Open(path)
	if err != nil {
		return asset, fmt.Errorf("loadAsset: %s", err)
	}
	defer file.Close()

	info, err := os.Stat(path)
	if err != nil {
		return asset, fmt.Errorf("loadAsset: %s", err)
	}

	if info.IsDir() {
		return asset, errors.New("loadAsset: File is a directory")
	}

	buffer, err := ioutil.ReadAll(file)
	if err != nil {
		return asset, fmt.Errorf("loadAsset: %s", err)
	}

	asset = Asset{Key: pathToProject(path)}
	if contentTypeFor(buffer) == "text" {
		asset.Value = string(buffer)
	} else {
		asset.Attachment = base64.StdEncoding.EncodeToString(buffer)
	}
	return asset, nil
}

func contentTypeFor(data []byte) string {
	contentType := http.DetectContentType(data)
	if strings.Contains(contentType, "text") {
		return "text"
	}
	return "binary"
}
