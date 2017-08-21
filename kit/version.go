package kit

import (
	"crypto"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/hashicorp/go-version"
	"github.com/inconshreveable/go-update"
)

var (
	// ThemeKitVersion is the version build of the library
	ThemeKitVersion, _ = version.NewVersion("0.7.1")
	// ThemeKitReleasesURL is the url that fetches all versions of themekit used for.
	// updating themekit. Change this for testing reasons.
	ThemeKitReleasesURL = "https://shopify-themekit.s3.amazonaws.com/releases/all.json"
	// ThemeKitLatestURL is the url that fetches new version of themekit. Change this
	// for testing reasons.
	ThemeKitLatestURL = "https://shopify-themekit.s3.amazonaws.com/releases/latest.json"
)

// IsNewUpdateAvailable will check if there is an update to the theme kit command
// and if there is one it will return true. Otherwise it will return false.
func IsNewUpdateAvailable() bool {
	release, err := FetchLatest()
	if err != nil {
		return false
	}
	return release.IsApplicable()
}

// InstallThemeKitVersion will take a semver string and parse it then check if that
// update is available and install it. If the string is 'latest' it will install
// the most current. If the string is latest and there is no update it will return an
// error. An error will also be returned if the requested version does not exist.
func InstallThemeKitVersion(ver string) error {
	if ver == "latest" {
		release, err := FetchLatest()
		if err != nil {
			return err
		}
		if !release.IsApplicable() {
			return fmt.Errorf("no applicable update available")
		}
		return applyUpdate(release.ForCurrentPlatform())
	}

	releases, err := FetchReleases()
	if err != nil {
		return err
	}
	requestedRelease := releases.Get(ver)
	if !requestedRelease.IsValid() {
		return fmt.Errorf("version %s not found", ver)
	}
	return applyUpdate(requestedRelease.ForCurrentPlatform())
}

// FetchLatest fetches the most recently released version of themekit.
//
// Used for internal purposes but exposed for cross package functionality
func FetchLatest() (Release, error) {
	var latest Release
	resp, err := http.Get(ThemeKitLatestURL)
	if err != nil {
		return latest, err
	}
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err == nil {
		err = json.Unmarshal(data, &latest)
	}
	return latest, err
}

// FetchReleases fetches all the versions of themekit ever released.
//
// Used for internal purposes but exposed for cross package functionality
func FetchReleases() (ReleasesList, error) {
	var releases ReleasesList
	resp, err := http.Get(ThemeKitReleasesURL)
	if err != nil {
		return releases, err
	}
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err == nil {
		err = json.Unmarshal(data, &releases)
	}
	return releases, err
}

func applyUpdate(platformRelease Platform) error {
	checksum, err := hex.DecodeString(platformRelease.Digest)
	if err != nil {
		return err
	}

	updateFile, err := http.Get(platformRelease.URL)
	if err != nil {
		return err
	}
	defer updateFile.Body.Close()

	err = update.Apply(updateFile.Body, update.Options{
		TargetPath: platformRelease.TargetPath, //used for testing
		Hash:       crypto.MD5,
		Checksum:   checksum,
	})

	if err != nil {
		if rerr := update.RollbackError(err); rerr != nil {
			return fmt.Errorf("Failed to rollback from bad update: %v", rerr)
		}
		return fmt.Errorf("Could not update and had to roll back. %v", err)
	}

	return nil
}
