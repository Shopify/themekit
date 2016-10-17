package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/Shopify/themekit/kit"
)

const (
	banner                 string = "----------------------------------------"
	updateAvailableMessage string = `
| An update for Theme Kit is available |
|                                      |
| To apply the update simply type      |
| the following command:               |
|                                      |
| theme update                         |`
)

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
	environments     kit.Environments
	themeClients     []kit.ThemeClient
	directory        string
	configPath       string
	environment      string
	allenvs          bool
	notifyFile       string
	password         string
	themeid          string
	domain           string
	bucketsize       int
	refillrate       int
	concurrency      int
	proxy            string
	timeout          time.Duration
	noUpdateNotifier bool
	ignoredFiles     flagArray
	ignores          flagArray

	bootstrapVersion string
	bootstrapPrefix  string
	setThemeID       bool
)

var ThemeCmd = &cobra.Command{
	Use:   "theme",
	Short: "Theme Kit is a tool kit for manipulating shopify themes",
	Long: `Theme Kit is a tool kit for manipulating shopify themes

Theme Kit is a Fast and cross platform tool that enables you
to build shopify themes with ease.

Complete documentation is available at http://themekit.cat`,
}

func init() {
	pwd, _ := os.Getwd()
	configPath = filepath.Join(pwd, "config.yml")

	ThemeCmd.PersistentFlags().StringVarP(&configPath, "config", "c", configPath, "path to config.yml")
	ThemeCmd.PersistentFlags().StringVarP(&environment, "env", "e", kit.DefaultEnvironment, "envionment to run the command")

	ThemeCmd.PersistentFlags().StringVarP(&directory, "dir", "d", "", "directory that command will take effect. (default current directory)")
	ThemeCmd.PersistentFlags().StringVar(&password, "password", "", "theme password. This will override what is in your config.yml")
	ThemeCmd.PersistentFlags().StringVar(&themeid, "themeid", "", "theme id. This will override what is in your config.yml")
	ThemeCmd.PersistentFlags().StringVar(&domain, "domain", "", "your shopify domain. This will override what is in your config.yml")
	ThemeCmd.PersistentFlags().IntVar(&bucketsize, "bucket", 0, "the bucket size for throttling. This will override what is in your config.yml")
	ThemeCmd.PersistentFlags().IntVar(&refillrate, "refill", 0, "the refill rate for throttling. This will override what is in your config.yml")
	ThemeCmd.PersistentFlags().IntVar(&concurrency, "concurrency", 0, "the refill rate for throttling. This will override what is in your config.yml")
	ThemeCmd.PersistentFlags().StringVar(&proxy, "proxy", "", "proxy for all theme requests. This will override what is in your config.yml")
	ThemeCmd.PersistentFlags().DurationVarP(&timeout, "timeout", "t", 0, "the timeout to kill any stalled processes. This will override what is in your config.yml")
	ThemeCmd.PersistentFlags().BoolVarP(&noUpdateNotifier, "no-update-notifier", "", false, "Stop theme kit from notifying about updates.")
	ThemeCmd.PersistentFlags().Var(&ignoredFiles, "ignored-file", "A single file to ignore, use the flag multiple times to add multiple.")
	ThemeCmd.PersistentFlags().Var(&ignores, "ignores", "A path to a file that contains ignore patterns.")

	watchCmd.Flags().StringVarP(&notifyFile, "notify", "n", "", "file to touch when workers have gone idle")
	watchCmd.Flags().BoolVarP(&allenvs, "allenvs", "a", false, "run command with all environments")
	removeCmd.Flags().BoolVarP(&allenvs, "allenvs", "a", false, "run command with all environments")
	replaceCmd.Flags().BoolVarP(&allenvs, "allenvs", "a", false, "run command with all environments")
	uploadCmd.Flags().BoolVarP(&allenvs, "allenvs", "a", false, "run command with all environments")

	bootstrapCmd.Flags().StringVar(&bootstrapVersion, "version", latestRelease, "version of Shopify Timber to use")
	bootstrapCmd.Flags().StringVar(&bootstrapPrefix, "prefix", "", "prefix to the Timber theme being created")

	ThemeCmd.AddCommand(bootstrapCmd, removeCmd, replaceCmd, uploadCmd, watchCmd, downloadCmd, versionCmd, updateCmd, configureCmd)
}

func initializeConfig(cmdName string, timesout bool) error {
	if cmdName != "update" && !noUpdateNotifier && isNewReleaseAvailable() {
		kit.Warnf("%s\n%s\n%s", banner, updateAvailableMessage, banner)
	}

	setFlagConfig()

	var err error
	if environments, err = kit.LoadEnvironments(configPath); err != nil {
		return err
	}

	themeClients = []kit.ThemeClient{}

	if allenvs {
		for env := range environments {
			config, err := environments.GetConfiguration(env)
			if err != nil {
				return err
			}
			client, err := kit.NewThemeClient(config)
			if err != nil {
				return err
			}
			themeClients = append(themeClients, client)
		}
	} else {
		config, err := environments.GetConfiguration(environment)
		if err != nil {
			return err
		}
		client, err := kit.NewThemeClient(config)
		if err != nil {
			return err
		}
		themeClients = []kit.ThemeClient{client}
	}

	return nil
}

func setFlagConfig() {
	kit.SetFlagConfig(kit.Configuration{
		Password:     password,
		ThemeID:      themeid,
		Domain:       domain,
		Directory:    directory,
		Proxy:        proxy,
		BucketSize:   bucketsize,
		RefillRate:   refillrate,
		Concurrency:  concurrency,
		IgnoredFiles: ignoredFiles.Value(),
		Ignores:      ignores.Value(),
		Timeout:      timeout,
	})
}
