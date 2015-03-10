package phoenix

import (
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

type Asset struct {
	Key        string `json:"key"`
	Value      string `json:"value,omitempty"`
	Attachment string `json:"attachment,omitempty"`
}

func (a Asset) String() string {
	return fmt.Sprintf("key: %s | value: %s | attachment: %s", a.Key, a.Value, a.Attachment)
}

func (a Asset) IsValid() bool {
	return len(a.Key) > 0 && (len(a.Value) > 0 || len(a.Attachment) > 0)
}

func LoadAsset(root, filename string) (asset Asset, err error) {
	path := toSlash(fmt.Sprintf("%s/%s", root, filename))
	file, err := os.Open(path)
	info, err := os.Stat(path)
	if err != nil {
		return
	}

	if info.IsDir() {
		err = errors.New("File is a directory")
		return
	}

	buffer := make([]byte, info.Size())
	_, err = file.Read(buffer)
	if err != nil {
		return
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
	} else {
		return "binary"
	}
}
