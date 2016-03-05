package theme

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

// Asset ... TODO
type Asset struct {
	Key        string `json:"key"`
	Value      string `json:"value,omitempty"`
	Attachment string `json:"attachment,omitempty"`
}

func (a Asset) String() string {
	return fmt.Sprintf("key: %s | value: %d bytes | attachment: %d bytes", a.Key, len([]byte(a.Value)), len([]byte(a.Attachment)))
}

// IsValid verifies that the Asset has a Key, and at least a Value or Attachment
func (a Asset) IsValid() bool {
	return len(a.Key) > 0 && (len(a.Value) > 0 || len(a.Attachment) > 0)
}

// Size ... TODO
func (a Asset) Size() int {
	if len(a.Value) > 0 {
		return len(a.Value)
	}
	return len(a.Attachment)
}

// ByAsset implements sort.Interface
type ByAsset []Asset

func (assets ByAsset) Len() int {
	return len(assets)
}

func (assets ByAsset) Swap(i, j int) {
	assets[i], assets[j] = assets[j], assets[i]
}

func (assets ByAsset) Less(i, j int) bool {
	return assets[i].Key < assets[j].Key
}

func findAllFiles(dir string) ([]string, error) {
	var files []string
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}

		files = append(files, path)
		return nil
	})

	return files, err
}

// LoadAssetsFromDirectory ... TODO
func LoadAssetsFromDirectory(dir string, ignore func(path string) bool) ([]Asset, error) {
	files, err := findAllFiles(dir)
	if err != nil {
		panic(err)
	}

	assets := []Asset{}
	for _, file := range files {
		assetKey, err := filepath.Rel(dir, file)
		if err != nil {
			panic(err)
		}
		if !ignore(assetKey) {
			asset, err := LoadAsset(dir, assetKey)
			if err == nil {
				assets = append(assets, asset)
			}
		}
	}

	return assets, nil
}

// LoadAsset ... TODO
func LoadAsset(root, filename string) (asset Asset, err error) {
	asset = Asset{}
	path := toSlash(fmt.Sprintf("%s/%s", root, filename))
	file, err := os.Open(path)
	if err != nil {
		return asset, fmt.Errorf("LoadAsset: %s", err)
	}
	defer file.Close()

	info, err := os.Stat(path)
	if err != nil {
		return asset, fmt.Errorf("LoadAsset: %s", err)
	}

	if info.IsDir() {
		err = errors.New("LoadAsset: File is a directory")
		return
	}

	buffer, err := ioutil.ReadAll(file)
	if err != nil {
		return asset, fmt.Errorf("LoadAsset: %s", err)
	}

	asset = Asset{Key: toSlash(filename)}
	if contentTypeFor(buffer) == "text" {
		asset.Value = string(buffer)
	} else {
		asset.Attachment = encode64(buffer)
	}
	return
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
