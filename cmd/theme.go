package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/spf13/cobra"
	"github.com/vbauerster/mpb"

	"github.com/Shopify/themekit/kit"
)

const updateAvailableMessage string = `
----------------------------------------
| An update for Theme Kit is available |
|                                      |
| To apply the update simply type      |
| the following command:               |
|                                      |
| theme update                         |
----------------------------------------
`

type flagArray struct {
	values []string
}

func (fa *flagArray) String() string {
	return strings.Join(fa.values, ",")
}

func (fa *flagArray) Set(value string) error {
	if len(value) > 0 {
		fa.values = append(fa.values, value)
	}
	return nil
}

func (fa *flagArray) Type() string {
	return "string"
}

func (fa *flagArray) Value() []string {
	if len(fa.values) == 0 {
		return nil
	}
	return fa.values
}

var (
	progress *mpb.Progress

	verbose          bool
	configPath       string
	allenvs          bool
	environment      string
	notifyFile       string
	noUpdateNotifier bool
	flagConfig       = kit.Configuration{}
	ignoredFiles     flagArray
	ignores          flagArray

	bootstrapVersion string
	bootstrapPrefix  string
	bootstrapURL     string
	bootstrapName    string
	setThemeID       bool
	openEdit         bool

	updateVersion string
)

// ThemeCmd is the main entry point to the theme kit command line interface.
var ThemeCmd = &cobra.Command{
	Use:   "theme",
	Short: "Theme Kit is a tool kit for manipulating shopify themes",
	Long: `Theme Kit is a tool kit for manipulating shopify themes

Theme Kit is a fast and cross platform tool that enables you
to build shopify themes with ease.

Complete documentation is available at https://shopify.github.io/themekit/`,
}

func init() {
	pwd, _ := os.Getwd()
	configPath = filepath.Join(pwd, "config.yml")

	ThemeCmd.PersistentFlags().StringVarP(&configPath, "config", "c", configPath, "path to config.yml")
	ThemeCmd.PersistentFlags().StringVarP(&environment, "env", "e", kit.DefaultEnvironment, "environment to run the command")
	ThemeCmd.PersistentFlags().StringVarP(&flagConfig.Directory, "dir", "d", "", "directory that command will take effect. (default current directory)")
	ThemeCmd.PersistentFlags().StringVarP(&flagConfig.Password, "password", "p", "", "theme password. This will override what is in your config.yml")
	ThemeCmd.PersistentFlags().StringVarP(&flagConfig.ThemeID, "themeid", "t", "", "theme id. This will override what is in your config.yml")
	ThemeCmd.PersistentFlags().StringVarP(&flagConfig.Domain, "store", "s", "", "your shopify domain. This will override what is in your config.yml")
	ThemeCmd.PersistentFlags().StringVar(&flagConfig.Proxy, "proxy", "", "proxy for all theme requests. This will override what is in your config.yml")
	ThemeCmd.PersistentFlags().DurationVar(&flagConfig.Timeout, "timeout", 0, "the timeout to kill any stalled processes. This will override what is in your config.yml")
	ThemeCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable more verbose output from the running command.")
	ThemeCmd.PersistentFlags().BoolVarP(&noUpdateNotifier, "no-update-notifier", "", false, "Stop theme kit from notifying about updates.")
	ThemeCmd.PersistentFlags().Var(&ignoredFiles, "ignored-file", "A single file to ignore, use the flag multiple times to add multiple.")
	ThemeCmd.PersistentFlags().Var(&ignores, "ignores", "A path to a file that contains ignore patterns.")

	watchCmd.Flags().StringVarP(&notifyFile, "notify", "n", "", "file to touch when workers have gone idle")
	watchCmd.Flags().BoolVarP(&allenvs, "allenvs", "a", false, "run command with all environments")
	removeCmd.Flags().BoolVarP(&allenvs, "allenvs", "a", false, "run command with all environments")
	replaceCmd.Flags().BoolVarP(&allenvs, "allenvs", "a", false, "run command with all environments")
	uploadCmd.Flags().BoolVarP(&allenvs, "allenvs", "a", false, "run command with all environments")
	openCmd.Flags().BoolVarP(&allenvs, "allenvs", "a", false, "run command with all environments")

	bootstrapCmd.Flags().StringVar(&bootstrapVersion, "version", latestRelease, "version of Shopify Timber to use")
	bootstrapCmd.Flags().StringVar(&bootstrapPrefix, "prefix", "", "prefix to the Timber theme being created")
	bootstrapCmd.Flags().StringVar(&bootstrapURL, "url", "", "a url to pull a project theme zip file from.")
	bootstrapCmd.Flags().StringVar(&bootstrapName, "name", "", "a name to define your theme on your shopify admin")

	updateCmd.Flags().StringVar(&updateVersion, "version", "latest", "version of themekit to install")

	openCmd.Flags().BoolVarP(&openEdit, "edit", "E", false, "open the web editor for the theme.")

	ThemeCmd.AddCommand(bootstrapCmd, removeCmd, replaceCmd, uploadCmd, watchCmd, downloadCmd, versionCmd, updateCmd, configureCmd, openCmd)
}

func generateThemeClients() ([]kit.ThemeClient, error) {
	progress = mpb.New(nil)

	themeClients := []kit.ThemeClient{}

	if !noUpdateNotifier && kit.IsNewUpdateAvailable() {
		kit.Print(kit.YellowText(updateAvailableMessage))
	}

	setFlagConfig()

	environments, err := kit.LoadEnvironments(configPath)
	if err != nil && os.IsNotExist(err) {
		return themeClients, fmt.Errorf("Could not file config file at %v", configPath)
	} else if err != nil {
		return themeClients, err
	}

	if !allenvs {
		config, err := environments.GetConfiguration(environment)
		if err != nil {
			return themeClients, err
		}
		environments = map[string]*kit.Configuration{environment: config}
	}

	for env := range environments {
		config, err := environments.GetConfiguration(env)
		if err != nil {
			return themeClients, err
		}
		client, err := kit.NewThemeClient(config)
		if err != nil {
			return themeClients, err
		}
		themeClients = append(themeClients, client)
	}

	return themeClients, nil
}

type cobraCommandE func(*cobra.Command, []string) error
type allEnvsCommand func(kit.ThemeClient, []string, *sync.WaitGroup)

func forEachClient(handler allEnvsCommand, callbacks ...allEnvsCommand) cobraCommandE {
	return func(cmd *cobra.Command, args []string) error {
		themeClients, err := generateThemeClients()
		defer progress.Stop()

		if err != nil {
			return err
		}
		wg := sync.WaitGroup{}
		for _, client := range themeClients {
			wg.Add(1)
			go handler(client, args, &wg)
		}
		wg.Wait()

		for _, client := range themeClients {
			for _, callback := range callbacks {
				callback(client, args, &wg)
			}
		}
		wg.Wait()

		return nil
	}
}

func setFlagConfig() {
	flagConfig.IgnoredFiles = ignoredFiles.Value()
	flagConfig.Ignores = ignores.Value()
	kit.SetFlagConfig(flagConfig)
}

func newProgressBar(count int, name string) *mpb.Bar {
	var bar *mpb.Bar
	if !verbose && progress != nil {
		bar = progress.AddBar(int64(count)).
			PrependName(fmt.Sprintf("[%s]: ", name), 0).
			AppendPercentage().
			PrependCounters(mpb.UnitNone, 0)
	}
	return bar
}

func incBar(bar *mpb.Bar) {
	if bar != nil {
		defer bar.Incr(1)
	}
}
