package cmd

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/Shopify/themekit/cmd/atom"
	"github.com/Shopify/themekit/kit"
)

const (
	masterBranch  = "master"
	latestRelease = "latest"
)

var (
	themeZipRoot   = "https://github.com/Shopify/Timber/archive/"
	timberFeedPath = "https://github.com/Shopify/Timber/releases.atom"
)

var bootstrapCmd = &cobra.Command{
	Use:   "bootstrap",
	Short: "Bootstrap a new theme using Shopify Timber",
	Long: `Bootstrap will download the latest release of Timber,
The most popular theme on Shopify. Bootstrap will also setup
your config file and create a new theme id for you.

For more documentation please see http://shopify.github.io/themekit/commands/#bootstrap
`,
	RunE: bootstrap,
}

func bootstrap(cmd *cobra.Command, args []string) error {
	zipLocation, err := getZipPath()
	if err != nil {
		return err
	}

	themeName := getThemeName()
	if arbiter.verbose {
		stdOut.Printf(
			"Attempting to create theme %s from %s",
			yellow(themeName),
			yellow(zipLocation),
		)
	}

	client, theme, err := kit.CreateTheme(themeName, zipLocation)
	if err != nil {
		return err
	}

	if err := saveConfiguration(client.Config); err != nil {
		return err
	}

	if arbiter.verbose {
		stdOut.Printf(
			"Successfully created theme '%s' with id of %s on shop %s",
			blue(theme.Name),
			blue(theme.ID),
			yellow(client.Config.Domain),
		)
	}

	if err := arbiter.generateThemeClients(nil, []string{}); err != nil {
		return err
	}
	return download(client, []string{})
}

func getZipPath() (string, error) {
	if bootstrapURL != "" {
		return bootstrapURL, nil
	}
	return zipPathForVersion(bootstrapVersion)
}

func getThemeName() string {
	if bootstrapName != "" {
		return bootstrapName
	}

	if bootstrapURL != "" {
		parts := strings.Split(bootstrapURL, "/")
		return bootstrapPrefix + strings.Replace(parts[len(parts)-1], ".zip", "", 1)
	}

	return bootstrapPrefix + "Timber-" + bootstrapVersion
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

	return fmt.Errorf(`invalid Timber Version: %s
Available Versions Are:
- %s`, version, strings.Join(entries, "\n- "))
}

func saveConfiguration(config *kit.Configuration) error {
	env, err := kit.LoadEnvironments(arbiter.configPath)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	env.SetConfiguration(kit.DefaultEnvironment, config)
	return env.Save(arbiter.configPath)
}
