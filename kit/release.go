package kit

import (
	"runtime"
	"sort"

	"github.com/hashicorp/go-version"
)

// Platform contains information for a release for a single architechture and
// operating system. It contains all the information needed to fetch it.
//
// Used for internal purposes but exposed for cross package functionality
type Platform struct {
	Name       string `json:"name"`
	URL        string `json:"url"`
	Digest     string `json:"digest"`
	TargetPath string `json:"TargetPath,omitempty"` // used for testing updating
}

// Release is a version of themekit released to the public.
//
// Used for internal purposes but exposed for cross package functionality
type Release struct {
	Version   string     `json:"version"`
	Platforms []Platform `json:"platforms"`
}

// IsValid returns true if there are any platforms for this release
func (r Release) IsValid() bool {
	return len(r.Platforms) > 0
}

// IsApplicable will return true if this release is an update to the current running
// version of themekit
func (r Release) IsApplicable() bool {
	version, err := version.NewVersion(r.Version)
	if err != nil {
		return false
	}
	return ThemeKitVersion.LessThan(version) && version.Metadata() == "" && version.Prerelease() == ""
}

// GetVersion will return the formatted version of this release.
func (r Release) GetVersion() *version.Version {
	version, _ := version.NewVersion(r.Version)
	return version
}

// ForCurrentPlatform will return the platform release for the current running
// operating system and arch
func (r Release) ForCurrentPlatform() Platform {
	platformKey := runtime.GOOS + "-" + runtime.GOARCH
	for _, p := range r.Platforms {
		if p.Name == platformKey {
			return p
		}
	}
	return Platform{}
}

// ReleasesList is a list of releases fetched from the server
type ReleasesList []Release

// Get will return the requested release by name. If no release is found, an
// invalid release will be returned.
func (releases ReleasesList) Get(ver string) Release {
	sort.Slice(releases, func(i, j int) bool {
		iversion, _ := version.NewVersion(releases[i].Version)
		jversion, _ := version.NewVersion(releases[j].Version)
		return jversion.LessThan(iversion)
	})

	if ver == "latest" {
		for _, release := range releases {
			releaseVersion, _ := version.NewVersion(release.Version)
			if releaseVersion.Metadata() == "" && releaseVersion.Prerelease() == "" {
				return release
			}
		}
	} else {
		requestedVersion, _ := version.NewVersion(ver)
		for _, release := range releases {
			releaseVersion, _ := version.NewVersion(release.Version)
			if requestedVersion.Equal(releaseVersion) {
				return release
			}
		}
	}
	return Release{}
}

// Del will find and remove a version from the list. The altered ReleaseList is
// returned
func (releases ReleasesList) Del(ver string) ReleasesList {
	requestedVersion, _ := version.NewVersion(ver)
	for i, r := range releases {
		releaseVersion, _ := version.NewVersion(r.Version)
		if requestedVersion.Equal(releaseVersion) {
			releases = append(releases[:i], releases[i+1:]...)
			return releases
		}
	}
	return releases
}
