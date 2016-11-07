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
	ThemeKitVersion, _ = version.NewVersion("0.5.1")
	releasesURL        = "https://shopify-themekit.s3.amazonaws.com/releases/all.json"
)

// LibraryInfo will return a string array with information about the library used
// for logging.
func LibraryInfo() string {
	messageSeparator := "\n----------------------------------------------------------------\n"
	info := fmt.Sprintf("\t%s %s", "ThemeKit - Shopify Theme Utilities", ThemeKitVersion.String())
	return fmt.Sprintf("%s%s%s", messageSeparator, info, messageSeparator)
}

// PrintInfo will output the version banner for the themekit library.
func PrintInfo() {
	LogNotify(LibraryInfo())
}

// IsNewUpdateAvailable will check if there is an update to the theme kit command
// and if there is one it will return true. Otherwise it will return false.
func IsNewUpdateAvailable() bool {
	list, err := fetchReleases()
	if err != nil {
		return false
	}
	return list.Get("latest").IsApplicable()
}

// InstallThemeKitVersion will take a semver string and parse it then check if that
// update is available and install it. If the string is 'latest' it will install
// the most current. If the string is latest and there is no update it will return an
// error. An error will also be returned if the requested version does not exist.
func InstallThemeKitVersion(ver string) error {
	releases, err := fetchReleases()
	if err != nil {
		return err
	}
	requestedRelease := releases.Get(ver)
	if !requestedRelease.IsValid() {
		return fmt.Errorf("Version %s not found.", ver)
	} else if ver == "latest" && !requestedRelease.IsApplicable() {
		return fmt.Errorf("No applicable update available.")
	}
	LogWarnf("Updating from %s to %s", ThemeKitVersion, requestedRelease.Version)
	err = applyUpdate(requestedRelease.ForCurrentPlatform())
	if err == nil {
		Printf(`
Successfully updated to theme kit version %v,
If you have troubles with this release please
report them to https://github.com/Shopify/themekit/issues
If your troubles are preventing you from working
you can roll back to the previous version using
the command 'theme update --version=v%s'
`, GreenText(requestedRelease.Version), YellowText(ThemeKitVersion))
	}
	return err
}

func fetchReleases() (releasesList, error) {
	var releases releasesList
	resp, err := http.Get(releasesURL)
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

func applyUpdate(platformRelease platform) error {
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
