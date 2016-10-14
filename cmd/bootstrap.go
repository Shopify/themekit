package cmd

import (
	"bytes"
	"errors"
	"net/http"

	"github.com/spf13/cobra"

	"github.com/Shopify/themekit/cmd/internal/atom"
	"github.com/Shopify/themekit/kit"
)

const (
	masterBranch   = "master"
	latestRelease  = "latest"
	themeZipRoot   = "https://github.com/Shopify/Timber/archive/"
	timberFeedPath = "https://github.com/Shopify/Timber/releases.atom"
)

var bootstrapCmd = &cobra.Command{
	Use:   "bootstrap",
	Short: "Bootstrap a new theme using Shopify Timber",
	Long: `Bootstrap will download the latest release of Timber,
The most popular theme on Shopify. Bootstrap will also setup
your config file and create a new theme id for you.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		setFlagConfig()

		zipLocation, err := zipPathForVersion(bootstrapVersion)
		if err != nil {
			return err
		}

		eventLog := make(chan kit.ThemeEvent)
		go consumeEventLog(eventLog, true, kit.DefaultTimeout)

		client := kit.CreateTheme(bootstrapPrefix+"Timber-"+bootstrapVersion, zipLocation, eventLog)
		if err := addConfiguration(client.GetConfiguration()); err != nil {
			return err
		}

		return download(client, []string{})
	},
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
	if version == latestRelease {
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
