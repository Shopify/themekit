package kit

import (
	"runtime"
	"sort"

	"github.com/hashicorp/go-version"
)

type platform struct {
	Name       string `json:"name"`
	URL        string `json:"url"`
	Digest     string `json:"digest"`
	TargetPath string `json:"TargetPath"` // used for testing updating
}

type release struct {
	Version   string     `json:"version"`
	Platforms []platform `json:"platforms"`
}

func (r release) IsValid() bool {
	return len(r.Platforms) > 0
}

func (r release) IsApplicable() bool {
	version, err := version.NewVersion(r.Version)
	if err != nil {
		return false
	}
	return ThemeKitVersion.LessThan(version) && version.Metadata() == "" && version.Prerelease() == ""
}

func (r release) GetVersion() *version.Version {
	version, _ := version.NewVersion(r.Version)
	return version
}

func (r release) ForCurrentPlatform() platform {
	platformKey := runtime.GOOS + "-" + runtime.GOARCH
	for _, p := range r.Platforms {
		if p.Name == platformKey {
			return p
		}
	}
	return platform{}
}

type releasesList []release

func (releases releasesList) Get(ver string) release {
	sort.Sort(releases)
	if ver == "latest" {
		return releases[0]
	}
	requestedVersion, _ := version.NewVersion(ver)
	for _, release := range releases {
		releaseVersion, _ := version.NewVersion(release.Version)
		if requestedVersion.Equal(releaseVersion) {
			return release
		}
	}
	return release{}
}

func (releases releasesList) Len() int {
	return len(releases)
}

func (releases releasesList) Swap(i, j int) {
	releases[i], releases[j] = releases[j], releases[i]
}

func (releases releasesList) Less(i, j int) bool {
	iversion, _ := version.NewVersion(releases[i].Version)
	jversion, _ := version.NewVersion(releases[j].Version)
	return jversion.LessThan(iversion)
}
