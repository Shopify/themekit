package commands

import (
	"bytes"
	"errors"
	"net/http"
	"os"

	"github.com/Shopify/themekit"
	"github.com/Shopify/themekit/atom"
)

const (
	MasterBranch   = "master"
	LatestRelease  = "latest"
	ThemeZipRoot   = "https://github.com/Shopify/Timber/archive/"
	TimberFeedPath = "https://github.com/Shopify/Timber/releases.atom"
)

type BootstrapOptions struct {
	BasicOptions
	Version     string
	Directory   string
	Environment string
	Prefix      string
	SetThemeID  bool
}

func BootstrapCommand(args map[string]interface{}) chan bool {
	options := BootstrapOptions{}

	extractString(&options.Version, "version", args)
	extractString(&options.Directory, "directory", args)
	extractString(&options.Environment, "environment", args)
	extractString(&options.Prefix, "prefix", args)
	extractBool(&options.SetThemeID, "setThemeId", args)
	extractThemeClient(&options.Client, args)
	extractEventLog(&options.EventLog, args)

	return Bootstrap(options)
}

func Bootstrap(options BootstrapOptions) chan bool {
	done := make(chan bool)
	go func() {
		doneCh := doBootstrap(options)
		done <- <-doneCh
	}()
	return done
}

func doBootstrap(options BootstrapOptions) chan bool {
	pwd, _ := os.Getwd()
	if pwd != options.Directory {
		os.Chdir(options.Directory)
	}

	zipLocation, err := zipPathForVersion(options.Version)
	if err != nil {
		themekit.NotifyError(err)
		done := make(chan bool)
		close(done)
		return done
	}

	name := "Timber-" + options.Version
	if len(options.Prefix) > 0 {
		name = options.Prefix + "-" + name
	}
	clientForNewTheme, themeEvents := options.Client.CreateTheme(name, zipLocation)
	mergeEvents(options.getEventLog(), []chan themekit.ThemeEvent{themeEvents})
	if options.SetThemeID {
		AddConfiguration(options.Directory, options.Environment, clientForNewTheme.GetConfiguration())
	}

	os.Chdir(pwd)

	downloadOptions := DownloadOptions{}
	downloadOptions.Client = clientForNewTheme
	downloadOptions.EventLog = options.getEventLog()

	done := Download(downloadOptions)

	return done
}

func zipPath(version string) string {
	return ThemeZipRoot + version + ".zip"
}

func zipPathForVersion(version string) (string, error) {
	if version == MasterBranch {
		return zipPath(MasterBranch), nil
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
	resp, err := http.Get(TimberFeedPath)
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
	buff.WriteString(themekit.RedText("Invalid Timber Version: " + version))
	buff.WriteString("\nAvailable Versions Are:")
	buff.WriteString("\n  - master")
	buff.WriteString("\n  - latest")
	for _, entry := range feed.Entries {
		buff.WriteString("\n  - " + entry.Title)
	}
	return errors.New(buff.String())
}
