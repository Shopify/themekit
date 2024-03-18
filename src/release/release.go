package release

import (
	"crypto"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"runtime"

	"github.com/hashicorp/go-version"
	binaryUpdate "github.com/inconshreveable/go-update"

	"github.com/Shopify/themekit/src/colors"
)

var (
	builds = map[string]string{
		"darwin-amd64":  "theme",
		"linux-386":     "theme",
		"linux-amd64":   "theme",
		"freebsd-386":   "theme",
		"freebsd-amd64": "theme",
		"windows-386":   "theme.exe",
		"windows-amd64": "theme.exe",
	}
	// ThemeKitVersion is the version build of the library
	ThemeKitVersion, _ = version.NewVersion("1.3.2")
)

const (
	releasesS3URL = "https://shopify-themekit.s3.amazonaws.com/releases/all.json"
	latestS3URL   = "https://shopify-themekit.s3.amazonaws.com/releases/latest.json"
)

type release struct {
	Version   string     `json:"version"`
	Platforms []platform `json:"platforms"`
}

func (r release) isValid() bool {
	return len(r.Platforms) > 0
}

func (r release) isApplicable() bool {
	version, err := version.NewVersion(r.Version)
	if err != nil {
		return false
	}
	return ThemeKitVersion.LessThan(version) && version.Metadata() == "" && version.Prerelease() == ""
}

func (r release) getVersion() *version.Version {
	version, _ := version.NewVersion(r.Version)
	return version
}

func (r release) forCurrentPlatform() platform {
	platformKey := runtime.GOOS + "-" + runtime.GOARCH
	for _, p := range r.Platforms {
		if p.Name == platformKey {
			return p
		}
	}
	return platform{}
}

// IsUpdateAvailable will check if there is an update to the theme kit command
// and if there is one it will return true. Otherwise it will return false.
func IsUpdateAvailable() bool {
	return checkUpdateAvailable(latestS3URL)
}

// Install will take a semver string and parse it then check if that
// update is available and install it. If the string is 'latest' it will install
// the most current. If the string is latest and there is no update it will return an
// error. An error will also be returned if the requested version does not exist.
func Install(ver string) error {
	installer := func(p platform) error {
		return applyUpdate(p, "")
	}
	if ver == "latest" {
		return installLatest(latestS3URL, installer)
	}
	return installVersion(ver, releasesS3URL, installer)
}

// Update will update the details of a release or deploy a new release with the
// deploy feed
func Update(key, secret, ver string, force bool) error {
	return update(ver, releasesS3URL, filepath.Join("build", "dist"), force, newS3Uploader(key, secret))
}

// Remove will remove a themekit release from the deployed releases list. This will
// prevent any users from installing the version again. This can only be done will
// appropriate S3 priviledges
func Remove(key, secret, ver string) error {
	return remove(ver, releasesS3URL, newS3Uploader(key, secret))
}

func checkUpdateAvailable(latestURL string) bool {
	release, err := fetchLatest(latestURL)
	if err != nil {
		return false
	}
	return release.isApplicable()
}

func installLatest(latestURL string, install func(platform) error) error {
	release, err := fetchLatest(latestURL)
	if err != nil {
		return err
	}
	if !release.isApplicable() {
		return fmt.Errorf("no applicable update available")
	}
	return install(release.forCurrentPlatform())
}

func installVersion(ver, releasesURL string, install func(platform) error) error {
	if _, err := version.NewVersion(ver); err != nil {
		return err
	}

	releases, err := fetchReleases(releasesURL)
	if err != nil {
		return err
	}
	requestedRelease := releases.get(ver)
	if !requestedRelease.isValid() {
		return fmt.Errorf("version %s not found", ver)
	}
	return install(requestedRelease.forCurrentPlatform())
}

func applyUpdate(platformRelease platform, targetPath string) error {
	checksum, err := hex.DecodeString(platformRelease.Digest)
	if err != nil {
		return err
	}

	updateFile, err := http.Get(platformRelease.URL)
	if err != nil {
		return err
	}
	defer updateFile.Body.Close()

	err = binaryUpdate.Apply(updateFile.Body, binaryUpdate.Options{
		TargetPath: targetPath,
		Hash:       crypto.MD5,
		Checksum:   checksum,
	})

	if err != nil {
		if rerr := binaryUpdate.RollbackError(err); rerr != nil {
			return fmt.Errorf("Failed to rollback from bad update: %v", rerr)
		}
		return fmt.Errorf("Could not update and had to roll back. %v", err)
	}

	return nil
}

func update(ver, releasesURL, distDir string, force bool, u uploader) error {
	if _, err := version.NewVersion(ver); err != nil {
		return err
	}

	if !force {
		requestedVersion, _ := version.NewVersion(ver)
		if !requestedVersion.Equal(ThemeKitVersion) {
			return errors.New("deploy version does not match themekit version")
		}
	}

	_, err := os.Stat(distDir)
	if os.IsNotExist(err) {
		return errors.New("Dist folder does not exist. Run 'make dist' before attempting to create a new release")
	}

	releases, err := fetchReleases(releasesURL)
	if err != nil {
		return err
	}

	if requestedRelease := releases.get(ver); !force && requestedRelease.isValid() {
		return errors.New("version has already been deployed")
	}

	newRelease, err := buildRelease(ver, distDir, u)
	if err != nil {
		return err
	}

	return updateDeploy(releases.add(newRelease), u)
}

func remove(ver, releaseURL string, u uploader) error {
	if _, err := version.NewVersion(ver); err != nil {
		return err
	}

	releases, err := fetchReleases(releaseURL)
	if err != nil {
		return err
	}

	requestedRelease := releases.get(ver)
	if !requestedRelease.isValid() {
		return errors.New("version has not be deployed")
	}

	return updateDeploy(releases.del(ver), u)
}

func updateDeploy(releases releasesList, u uploader) error {
	colors.ColorStdOut.Printf("Updating releases")
	if err := u.JSON("releases/all.json", releases); err != nil {
		return err
	}
	return u.JSON("releases/latest.json", releases.get("latest"))
}

func fetchLatest(url string) (release, error) {
	var latest release
	resp, err := http.Get(url)
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

func buildRelease(ver, distDir string, u uploader) (release, error) {
	colors.ColorStdOut.Printf("Building %s", colors.Green(ver))
	newRelease := release{Version: ver, Platforms: []platform{}}

	for platformName, binName := range builds {
		plat, err := buildPlatform(ver, platformName, distDir, binName, u)
		if err != nil {
			return newRelease, err
		}
		newRelease.Platforms = append(newRelease.Platforms, plat)
	}

	return newRelease, nil
}
