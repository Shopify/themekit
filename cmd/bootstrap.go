package cmd

import (
	"fmt"
	"net/http"
	"strings"

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
		return bootstrap()
	},
}

func bootstrap() error {
	zipLocation, err := zipPathForVersion(bootstrapVersion)
	if err != nil {
		return err
	}

	themeName := bootstrapPrefix + "Timber-" + bootstrapVersion
	kit.Printf(
		"Attempting to create theme %s from %s",
		kit.YellowText(themeName),
		kit.YellowText(zipLocation),
	)

	client, theme, err := kit.CreateTheme(themeName, zipLocation)
	if err != nil {
		return err
	}

	if err := saveConfiguration(client.Config); err != nil {
		return err
	}

	kit.Printf(
		"Successfully created theme '%s' with id of %s on shop %s",
		kit.BlueText(theme.Name),
		kit.BlueText(theme.ID),
		kit.YellowText(client.Config.Domain),
	)

	return download(client, []string{})
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

func zipPath(version string) string {
	return themeZipRoot + version + ".zip"
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
	entries := []string{"master", "latest"}

	for _, entry := range feed.Entries {
		entries = append(entries, entry.Title)
	}

	return fmt.Errorf(`Invalid Timber Version: %s
Available Versions Are:
- %s
`, version, strings.Join(entries, "\n- "))
}
