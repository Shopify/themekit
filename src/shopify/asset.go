package shopify

import (
	"bytes"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/Shopify/themekit/src/env"
	"github.com/Shopify/themekit/src/file"
)

// Asset represents an asset from the shopify server.
type Asset struct {
	Key         string `json:"key"`
	Value       string `json:"value,omitempty"`
	Attachment  string `json:"attachment,omitempty"`
	ContentType string `json:"content_type,omitempty"`
	ThemeID     int64  `json:"theme_id,omitempty"`
	UpdatedAt   string `json:"updated_at,omitempty"`
	Checksum    string `json:"checksum"`
}

var (
	// ErrAssetIsDir is the error returned if you try and load a directory with ReadAsset
	ErrAssetIsDir = errors.New("requested asset is a directory")
)

// ReadAsset will read a single asset from disk
func ReadAsset(e *env.Env, filename string) (Asset, error) {
	return readAsset(e.Directory, filename)
}

// FindAssets will load all assets for paths passed in, this also means that it will
// read directories recursively. If no paths are passed in then the whole project
// directory will be read
func FindAssets(e *env.Env, paths ...string) (assets map[string]string, err error) {
	filter, err := file.NewFilter(e.Directory, e.IgnoredFiles, e.Ignores)
	if err != nil {
		return map[string]string{}, err
	}

	if len(paths) == 0 {
		return loadAssetsFromDirectory(e.Directory, "", filter.Match)
	}

	for _, path := range paths {
		asset, err := readAsset(e.Directory, path)
		if err == ErrAssetIsDir {
			dirAssets, err := loadAssetsFromDirectory(e.Directory, path, filter.Match)
			if err != nil {
				return map[string]string{}, err
			}
			for filename, checksum := range dirAssets {
				assets[filename] = checksum
			}
		} else if err != nil {
			return map[string]string{}, err
		} else if !filter.Match(asset.Key) {
			assets[asset.Key] = asset.Checksum
		}
	}

	return assets, nil
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

	contents, err := asset.contents()
	if err != nil {
		return err
	}

	_, err = file.Write(contents)
	return err
}

func (asset Asset) contents() ([]byte, error) {
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

func assetsToFilenames(assets []Asset) []string {
	filenames := []string{}
	for _, asset := range assets {
		filenames = append(filenames, asset.Key)
	}
	return filenames
}

func loadAssetsFromDirectory(root, dir string, ignore func(path string) bool) (map[string]string, error) {
	assets := map[string]string{}
	return assets, filepath.Walk(filepath.Join(root, dir), func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}

		key, sum, err := readAssetInfo(root, relPath)
		if err == ErrAssetIsDir {
			return nil
		} else if err != nil {
			return err
		}

		if !ignore(key) {
			assets[key] = sum
		}
		return nil
	})
}

func readAssetInfo(root, filename string) (string, string, error) {
	path := filepath.Join(root, filename)

	key, err := filepath.Rel(root, path)
	if err != nil {
		return "", "", err
	}

	key = filepath.ToSlash(key)

	info, err := os.Stat(path)
	if err != nil {
		return "", "", fmt.Errorf("readAssetInfo: %s", err)
	}

	if info.IsDir() {
		return "", "", ErrAssetIsDir
	}

	checksum, err := calculateAssetChecksum(path)
	return key, checksum, err
}

func calculateAssetChecksum(path string) (string, error) {
	buffer, err := ioutil.ReadFile(path)
	if err != nil {
		return "", err
	}
	if filepath.Ext(path) == ".json" {
		var out bytes.Buffer
		if err := json.Compact(&out, buffer); err != nil {
			return "", err
		}
		buffer = out.Bytes()
	}
	return fmt.Sprintf("%x", sha1.Sum(buffer)), err
}

func readAsset(root, filename string) (asset Asset, err error) {
	key, sum, err := readAssetInfo(root, filename)
	if err != nil {
		return Asset{}, err
	}

	asset = Asset{Key: filepath.ToSlash(key), Checksum: sum}

	buffer, err := ioutil.ReadFile(filepath.Join(root, filename))
	if err != nil {
		return Asset{}, fmt.Errorf("readAsset: %s", err)
	}

	contentType := http.DetectContentType(buffer)
	if strings.Contains(contentType, "text") {
		asset.Value = string(buffer)
	} else {
		asset.Attachment = base64.StdEncoding.EncodeToString(buffer)
	}
	return asset, nil
}
