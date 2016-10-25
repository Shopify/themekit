package kit

import (
	"encoding/base64"
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
func (a Asset) IsValid() bool {
	return len(a.Key) > 0 && (len(a.Value) > 0 || len(a.Attachment) > 0)
}

// Size will return the length of the value or attachment depending on which exists.
func (a Asset) Size() int {
	if len(a.Value) > 0 {
		return len(a.Value)
	}
	return len(a.Attachment)
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
	path := toSlash(fmt.Sprintf("%s/%s", root, filename))
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

	asset = Asset{Key: toSlash(filename)}
	if contentTypeFor(buffer) == "text" {
		asset.Value = string(buffer)
	} else {
		asset.Attachment = encode64(buffer)
	}
	return asset, nil
}

func toSlash(path string) string {
	newpath := filepath.ToSlash(path)
	if strings.Index(newpath, "\\") >= 0 {
		newpath = strings.Replace(newpath, "\\", "/", -1)
	}
	return newpath
}

func encode64(data []byte) string {
	return base64.StdEncoding.EncodeToString(data)
}

func contentTypeFor(data []byte) string {
	contentType := http.DetectContentType(data)
	if strings.Contains(contentType, "text") {
		return "text"
	}
	return "binary"
}
