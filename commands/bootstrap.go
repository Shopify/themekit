package commands

import (
	"bytes"
	"errors"
	"net/http"
	"os"

	"github.com/Shopify/themekit/atom"
	"github.com/Shopify/themekit/kit"
)

const (
	masterBranch = "master"
	// LatestRelease (github's latest release)
	LatestRelease  = "latest"
	themeZipRoot   = "https://github.com/Shopify/Timber/archive/"
	timberFeedPath = "https://github.com/Shopify/Timber/releases.atom"
)

// BootstrapCommand bootstraps a new theme using Shopify Timber
func BootstrapCommand(args Args, done chan bool) {
	pwd, _ := os.Getwd()
	if pwd != args.Directory {
		os.Chdir(args.Directory)
	}

	zipLocation, err := zipPathForVersion(args.Version)
	if err != nil {
		kit.NotifyError(err)
		close(done)
	}

	name := "Timber-" + args.Version
	if len(args.Prefix) > 0 {
		name = args.Prefix + "-" + name
	}
	clientForNewTheme := args.ThemeClient.CreateTheme(name, zipLocation)
	if args.SetThemeID {
		AddConfiguration(args.Directory, args.Environment, clientForNewTheme.GetConfiguration())
	}

	os.Chdir(pwd)

	downloadOptions := Args{}
	downloadOptions.ThemeClient = clientForNewTheme

	DownloadCommand(downloadOptions, done)
}

func zipPath(version string) string {
	return themeZipRoot + version + ".zip"
}

func zipPathForVersion(version string) (string, error) {
	if version == masterBranch {
		return zipPath(masterBranch), nil
	}

	feed, err := downloadAtomFeed()
	if err != nil {
		return "", err
	}

	entry, err := findReleaseWith(feed, version)
	if err != nil {
		return "", err
	}

	return zipPath(entry.Title), nil
}

func downloadAtomFeed() (atom.Feed, error) {
	resp, err := http.Get(timberFeedPath)
	if err != nil {
		return atom.Feed{}, err
	}
	defer resp.Body.Close()

	feed, err := atom.LoadFeed(resp.Body)
	if err != nil {
		return atom.Feed{}, err
	}
	return feed, nil
}

func findReleaseWith(feed atom.Feed, version string) (atom.Entry, error) {
	if version == LatestRelease {
		return feed.LatestEntry(), nil
	}
	for _, entry := range feed.Entries {
		if entry.Title == version {
			return entry, nil
		}
	}
	return atom.Entry{Title: "Invalid Feed"}, buildInvalidVersionError(feed, version)
}

func buildInvalidVersionError(feed atom.Feed, version string) error {
	buff := bytes.NewBuffer([]byte{})
	buff.WriteString(kit.RedText("Invalid Timber Version: " + version))
	buff.WriteString("\nAvailable Versions Are:")
	buff.WriteString("\n  - master")
	buff.WriteString("\n  - latest")
	for _, entry := range feed.Entries {
		buff.WriteString("\n  - " + entry.Title)
	}
	return errors.New(buff.String())
}
