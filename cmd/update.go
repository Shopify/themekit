package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"runtime"

	"github.com/spf13/cobra"

	"github.com/Shopify/themekit/kit"
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

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update Theme kit to the newest verion.",
	Long: `Update will check for a new release, then
if there is an applicable update it will
download it and apply it.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := initializeConfig(cmd.Name(), true); err != nil {
			return err
		}

		latestRelease, err := downloadReleaseForPlatform()
		if err == nil {
			if latestRelease.IsApplicable() {
				kit.Warnf("Updating from", kit.ThemeKitVersion, "to", kit.ParseVersionString(latestRelease.Version))
				releaseForPlatform := findAppropriateRelease(latestRelease)
				return kit.ApplyUpdate(releaseForPlatform.URL, releaseForPlatform.Digest)
			} else {
				return fmt.Errorf("No applicable update available.")
			}
		}
		return err
	},
}

func (r release) IsApplicable() bool {
	return kit.ThemeKitVersion.Compare(kit.ParseVersionString(r.Version)) == kit.VersionLessThan
}

func isNewReleaseAvailable() bool {
	latestRelease, err := downloadReleaseForPlatform()
	if err != nil {
		return false
	}
	return latestRelease.IsApplicable()
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
