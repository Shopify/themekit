package release

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"sort"

	"github.com/hashicorp/go-version"
)

type releasesList []release

func fetchReleases(url string) (releasesList, error) {
	var releases releasesList
	resp, err := http.Get(url)
	if err != nil {
		return releases, err
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return releases, err
	}

	return releases, json.Unmarshal(data, &releases)
}

func (releases releasesList) add(newRelease release) releasesList {
	return append(releases.del(newRelease.Version), newRelease)
}

func (releases releasesList) get(ver string) release {
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
	return release{}
}

func (releases releasesList) del(ver string) releasesList {
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
