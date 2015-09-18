package commands

import (
	"encoding/json"
	"fmt"
	"github.com/Shopify/themekit"
	"io"
	"io/ioutil"
	"net/http"
	"runtime"
)

const LatestReleasesUrl string = "https://shopify-themekit.s3.amazonaws.com/releases/latest.json"

type platform struct {
	Name   string `json:"name"`
	URL    string `json:"url"`
	Digest string `json:"digest"`
}

type release struct {
	Version   string     `json:"version"`
	Platforms []platform `json:"platforms"`
}

func Version(r release) themekit.Version {
	return themekit.ParseVersionString(r.Version)
}

func (r release) IsApplicable() bool {
	return themekit.TKVersion.Compare(Version(r)) == themekit.VersionLessThan
}

func IsNewReleaseAvailable() bool {
	latestRelease, err := downloadReleaseForPlatform()
	if err != nil {
		return false
	}
	return latestRelease.IsApplicable()
}

func UpdateCommand(args map[string]interface{}) chan bool {
	latestRelease, err := downloadReleaseForPlatform()
	if err == nil {
		if latestRelease.IsApplicable() {
			fmt.Println("Updating from", themekit.TKVersion, "to", Version(latestRelease))
			releaseForPlatform := findAppropriateRelease(latestRelease)
			themekit.ApplyUpdate(releaseForPlatform.URL, releaseForPlatform.Digest)
		}
	}
	res := make(chan bool)
	close(res)
	return res
}

func downloadReleaseForPlatform() (release, error) {
	resp, err := http.Get(LatestReleasesUrl)
	if err != nil {
		return release{}, err
	}
	defer resp.Body.Close()
	return parseManifest(resp.Body)
}

func parseManifest(r io.Reader) (release, error) {
	var parsed release
	data, err := ioutil.ReadAll(r)
	if err == nil {
		err = json.Unmarshal(data, &parsed)
	}
	return parsed, err
}

func findAppropriateRelease(r release) platform {
	platformKey := runtime.GOOS + "-" + runtime.GOARCH
	for _, p := range r.Platforms {
		if p.Name == platformKey {
			return p
		}
	}
	return platform{}
}
