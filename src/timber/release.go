package timber

import (
	"bytes"
	"fmt"
	"net/http"
	"text/template"

	"github.com/Shopify/themekit/src/atom"
)

const (
	timberFeedPath = "https://github.com/Shopify/Timber/releases.atom"
)

var (
	invalidVersionTmplt = template.Must(template.New("invalidVersionError").Parse(`Invalid Timber Version: {{ .Requested }}
  Available Versions Are:
  - master
  - latest
  {{- range .Versions }}
  - {{ . }}
  {{- end }}`))
)

// GetVersionPath will return the download path for a requested version
func GetVersionPath(version string) (string, error) {
	return getVersionPath(version, timberFeedPath)
}

func getVersionPath(version, feedPath string) (string, error) {
	if version == "master" {
		return "https://github.com/Shopify/Timber/archive/master.zip", nil
	}

	feed, err := downloadThemeReleaseAtomFeed(feedPath)
	if err != nil {
		return "", err
	}

	entry, err := findThemeReleaseWith(feed, version)
	if err != nil {
		return "", err
	}

	return "https://github.com/Shopify/Timber/archive/" + entry.Title + ".zip", nil
}

func downloadThemeReleaseAtomFeed(feedPath string) (atom.Feed, error) {
	resp, err := http.Get(feedPath)
	if err != nil {
		return atom.Feed{}, err
	}
	defer resp.Body.Close()

	feed, err := atom.LoadFeed(resp.Body)
	if err != nil {
		return atom.Feed{}, err
	}

	return feed, resp.Body.Close()
}

func findThemeReleaseWith(feed atom.Feed, version string) (atom.Entry, error) {
	if version == "latest" {
		return feed.LatestEntry(), nil
	}

	entries := []string{}
	for _, entry := range feed.Entries {
		if entry.Title == version {
			return entry, nil
		}
		entries = append(entries, entry.Title)
	}

	var tpl bytes.Buffer
	invalidVersionTmplt.Execute(&tpl, struct {
		Requested string
		Versions  []string
	}{version, entries})

	return atom.Entry{Title: "Invalid Feed"}, fmt.Errorf(tpl.String())
}
