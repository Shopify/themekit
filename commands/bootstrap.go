package commands

import (
	"bytes"
	"errors"
	"github.com/csaunders/phoenix"
	"net/http"
	"os"
)

const (
	MasterBranch   = "master"
	LatestRelease  = "latest"
	ThemeZipRoot   = "https://github.com/Shopify/Timber/archive/"
	TimberFeedPath = "https://github.com/Shopify/Timber/releases.atom"
)

func BootstrapCommand(args map[string]interface{}) chan bool {
	var client phoenix.ThemeClient
	var version, dir, env, prefix string
	var setThemeId bool
	extractString(&version, "version", args)
	extractString(&dir, "directory", args)
	extractString(&env, "environment", args)
	extractString(&prefix, "prefix", args)
	extractBool(&setThemeId, "setThemeId", args)
	extractThemeClient(&client, args)

	return Bootstrap(client, prefix, version, dir, env, setThemeId)
}

func Bootstrap(client phoenix.ThemeClient, prefix, version, directory, environment string, setThemeId bool) chan bool {
	done := make(chan bool)
	eventLog := make(chan phoenix.ThemeEvent)
	go func() {
		doneCh := doBootstrap(client, prefix, version, directory, environment, setThemeId, eventLog)
		done <- <-doneCh
	}()
	return done
}

func doBootstrap(client phoenix.ThemeClient, prefix, version, directory, environment string, setThemeId bool, eventLog chan phoenix.ThemeEvent) chan bool {
	pwd, _ := os.Getwd()
	if pwd != directory {
		os.Chdir(directory)
	}

	zipLocation, err := zipPathForVersion(version)
	if err != nil {
		phoenix.NotifyError(err)
		done := make(chan bool)
		close(done)
		close(eventLog)
		return done
	}

	name := "Timber-" + version
	if len(prefix) > 0 {
		name = prefix + "-" + name
	}
	clientForNewTheme, themeEvents := client.CreateTheme(name, zipLocation)
	mergeEvents(eventLog, []chan phoenix.ThemeEvent{themeEvents})
	if setThemeId {
		AddConfiguration(directory, environment, clientForNewTheme.GetConfiguration())
	}

	os.Chdir(pwd)
	done := Download(clientForNewTheme, []string{})

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

func downloadAtomFeed() (phoenix.Feed, error) {
	resp, err := http.Get(TimberFeedPath)
	if err != nil {
		return phoenix.Feed{}, err
	}
	defer resp.Body.Close()

	feed, err := phoenix.LoadFeed(resp.Body)
	if err != nil {
		return phoenix.Feed{}, err
	}
	return feed, nil
}

func findReleaseWith(feed phoenix.Feed, version string) (phoenix.Entry, error) {
	if version == LatestRelease {
		return feed.LatestEntry(), nil
	}
	for _, entry := range feed.Entries {
		if entry.Title == version {
			return entry, nil
		}
	}
	return phoenix.Entry{Title: "Invalid Feed"}, buildInvalidVersionError(feed, version)
}

func buildInvalidVersionError(feed phoenix.Feed, version string) error {
	buff := bytes.NewBuffer([]byte{})
	buff.Write([]byte(phoenix.RedText("Invalid Timber Version: " + version)))
	buff.Write([]byte("\nAvailable Versions Are:"))
	buff.Write([]byte("\n  - master"))
	buff.Write([]byte("\n  - latest"))
	for _, entry := range feed.Entries {
		buff.Write([]byte("\n  - " + entry.Title))
	}
	return errors.New(buff.String())
}
