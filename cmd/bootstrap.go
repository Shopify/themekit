package cmd

import (
	"bytes"
	"fmt"
	"net/http"
	"os"
	"strings"
	"text/template"

	"github.com/spf13/cobra"

	"github.com/Shopify/themekit/cmd/atom"
	"github.com/Shopify/themekit/kit"
)

const (
	masterBranch  = "master"
	latestRelease = "latest"
)

var (
	themeZipRoot        = "https://github.com/Shopify/Timber/archive/"
	timberFeedPath      = "https://github.com/Shopify/Timber/releases.atom"
	invalidVersionTmplt = template.Must(template.New("invalidVersionError").Parse(`Invalid Timber Version: {{ .Requested }}
Available Versions Are:
- master
- latest
{{- range .Versions }}
- {{ . }}
{{- end }}`))
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
	zipLocation, err := getNewThemeZipPath()
	if err != nil {
		return err
	}

	themeName := getNewThemeName()
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

func getNewThemeZipPath() (string, error) {
	if bootstrapURL != "" {
		return bootstrapURL, nil
	} else if bootstrapVersion == masterBranch {
		return themeZipRoot + masterBranch + ".zip", nil
	}

	feed, err := downloadThemeReleaseAtomFeed()
	if err != nil {
		return "", err
	}

	entry, err := findThemeReleaseWith(feed, bootstrapVersion)
	if err != nil {
		return "", err
	}

	return themeZipRoot + entry.Title + ".zip", nil
}

func getNewThemeName() string {
	if bootstrapName != "" {
		return bootstrapName
	}

	if bootstrapURL != "" {
		parts := strings.Split(bootstrapURL, "/")
		return bootstrapPrefix + strings.Replace(parts[len(parts)-1], ".zip", "", 1)
	}

	return bootstrapPrefix + "Timber-" + bootstrapVersion
}

func downloadThemeReleaseAtomFeed() (atom.Feed, error) {
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

func findThemeReleaseWith(feed atom.Feed, version string) (atom.Entry, error) {
	if version == latestRelease {
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

func saveConfiguration(config *kit.Configuration) error {
	env, err := kit.LoadEnvironments(arbiter.configPath)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	flagEnvs := arbiter.environments.Value()
	if len(flagEnvs) == 0 {
		env.SetConfiguration(kit.DefaultEnvironment, config)
	} else {
		for _, envName := range flagEnvs {
			env.SetConfiguration(envName, config)
		}
	}

	return env.Save(arbiter.configPath)
}
