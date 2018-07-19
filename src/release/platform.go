package release

import (
	"crypto/md5"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

type platform struct {
	Name   string `json:"name"`
	URL    string `json:"url"`
	Digest string `json:"digest"`
}

func buildPlatform(ver, platformName, distDir, binName string, u uploader, platformChan chan platform) error {
	f, err := os.Open(filepath.Join(distDir, platformName, binName))
	if err != nil {
		return err
	}
	defer f.Close()

	data, err := ioutil.ReadAll(f)
	if err != nil {
		return err
	}

	fullName := strings.Join([]string{ver, platformName, binName}, "/")
	url, err := u.File(fullName, f)
	if err != nil {
		return err
	}

	platformChan <- platform{
		Name:   platformName,
		URL:    url,
		Digest: fmt.Sprintf("%x", md5.Sum(data)),
	}
	return nil
}
