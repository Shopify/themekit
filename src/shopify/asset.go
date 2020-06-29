package shopify

import (
	"bytes"
	"crypto/md5"
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
	Checksum    string `json:"checksum,omitempty"`
	UpdatedAt   string `json:"updated_at,omitempty"`
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
func FindAssets(e *env.Env, paths ...string) (assets []Asset, err error) {
	filter, err := file.NewFilter(e.Directory, e.IgnoredFiles, e.Ignores)
	if err != nil {
		return []Asset{}, err
	}

	if len(paths) == 0 {
		return loadAssetsFromDirectory(e, "", filter.Match)
	}

	for _, path := range paths {
		asset, err := readAsset(e.Directory, path)
		if err == ErrAssetIsDir {
			dirAssets, err := loadAssetsFromDirectory(e, path, filter.Match)
			if err != nil {
				return []Asset{}, err
			}
			assets = append(assets, dirAssets...)
		} else if err != nil {
			return []Asset{}, err
		} else if !filter.Match(asset.Key) {
			assets = append(assets, asset)
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

func loadAssetsFromDirectory(e *env.Env, dir string, ignore func(path string) bool) (assets []Asset, err error) {
	var root = e.Directory
	err = filepath.Walk(filepath.Join(root, dir), func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		assetKey, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		assetKey = filepath.ToSlash(assetKey)
		if !ignore(assetKey) {
			var asset, _ = ReadAsset(e, assetKey) // TODO handle error
			assets = append(assets, asset)
		}
		return nil
	})
	return
}

func readAsset(root, filename string) (asset Asset, err error) {
	path := filepath.Join(root, filename)

	key, err := filepath.Rel(root, path)
	if err != nil {
		return Asset{}, err
	}

	asset = Asset{Key: filepath.ToSlash(key)}
	file, err := os.Open(path)
	if err != nil {
		return Asset{}, fmt.Errorf("readAsset: %s", err)
	}
	defer file.Close()

	info, err := os.Stat(path)
	if err != nil {
		return Asset{}, fmt.Errorf("readAsset: %s", err)
	}

	if info.IsDir() {
		return Asset{}, ErrAssetIsDir
	}

	buffer, err := ioutil.ReadAll(file)
	if err != nil {
		return Asset{}, fmt.Errorf("readAsset: %s", err)
	}

	contentType := http.DetectContentType(buffer)
	if strings.Contains(contentType, "text") {
		asset.Value = string(buffer)
		asset.Checksum = calculateTextChecksum(asset.Value, filepath.Ext(asset.Key) == ".json")
	} else {
		asset.Attachment = base64.StdEncoding.EncodeToString(buffer)
		asset.Checksum = calculateByteArrayChecksum(buffer)
	}
	return asset, nil
}

func calculateTextChecksum(value string, isJSON bool) (checksum string) {
	if isJSON {
		buf := new(bytes.Buffer)
		json.Compact(buf, []byte(value))
		return fmt.Sprintf("%x", md5.Sum(buf.Bytes()))
	}
	return fmt.Sprintf("%x", md5.Sum([]byte(value)))
}

func calculateByteArrayChecksum(value []byte) (checksum string) {
	hash := md5.New()
	hash.Write(value)
	return fmt.Sprintf("%x", hash.Sum(nil))
}
