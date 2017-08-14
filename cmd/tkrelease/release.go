package main

import (
	"crypto/md5"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/Shopify/themekit/kit"
)

var (
	builds = map[string]string{
		"darwin-amd64":  "theme",
		"darwin-386":    "theme",
		"linux-386":     "theme",
		"linux-amd64":   "theme",
		"windows-386":   "theme.exe",
		"windows-amd64": "theme.exe",
	}
)

func buildRelease(ver string, uploader *s3Uploader) (kit.Release, error) {
	var wg sync.WaitGroup
	wg.Add(len(builds))
	newRelease := kit.Release{Version: ver, Platforms: []kit.Platform{}}
	finished := make(chan bool, 1)
	errChan := make(chan error)
	platformChan := make(chan kit.Platform)

	for platformName, binName := range builds {
		go buildPlatform(ver, platformName, binName, uploader, platformChan, errChan, &wg)
	}

	go func() {
		wg.Wait()
		close(finished)
	}()

	for {
		select {
		case <-finished:
			return newRelease, nil
		case platform := <-platformChan:
			newRelease.Platforms = append(newRelease.Platforms, platform)
		case err := <-errChan:
			return kit.Release{}, err
		}
	}
}

func buildPlatform(ver, platformName, binName string, uploader *s3Uploader, platformChan chan kit.Platform, errChan chan error, wg *sync.WaitGroup) {
	defer wg.Done()

	f, err := os.Open(filepath.Join(distDir, platformName, binName))
	if err != nil {
		errChan <- err
		return
	}
	defer f.Close()

	data, err := ioutil.ReadAll(f)
	if err != nil {
		errChan <- err
		return
	}

	fullName := strings.Join([]string{ver, platformName, binName}, "/")
	url, err := uploader.file(fullName, f)
	if err != nil {
		errChan <- err
		return
	}

	platformChan <- kit.Platform{
		Name:   platformName,
		URL:    url,
		Digest: fmt.Sprintf("%x", md5.Sum(data)),
	}
}
