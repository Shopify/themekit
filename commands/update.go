package commands

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"runtime"

	"github.com/Shopify/themekit"
)

const latestReleasesURL string = "https://shopify-themekit.s3.amazonaws.com/releases/latest.json"

type platform struct {
	Name   string `json:"name"`
	URL    string `json:"url"`
	Digest string `json:"digest"`
}

type release struct {
	Version   string     `json:"version"`
	Platforms []platform `json:"platforms"`
}

// Version ... TODO
func Version(r release) themekit.Version {
	return themekit.ParseVersionString(r.Version)
}

// IsApplicable ... TODO
func (r release) IsApplicable() bool {
	return themekit.TKVersion.Compare(Version(r)) == themekit.VersionLessThan
}

// IsNewReleaseAvailable ... TODO
func IsNewReleaseAvailable() bool {
	latestRelease, err := downloadReleaseForPlatform()
	if err != nil {
		return false
	}
	return latestRelease.IsApplicable()
}

// UpdateCommand ... TODO
func UpdateCommand(args Args, done chan bool) {
	latestRelease, err := downloadReleaseForPlatform()
	if err == nil {
		if latestRelease.IsApplicable() {
			fmt.Println("Updating from", themekit.TKVersion, "to", Version(latestRelease))
			releaseForPlatform := findAppropriateRelease(latestRelease)
			themekit.ApplyUpdate(releaseForPlatform.URL, releaseForPlatform.Digest)
		}
	}
	close(done)
}

func downloadReleaseForPlatform() (release, error) {
	resp, err := http.Get(latestReleasesURL)
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
