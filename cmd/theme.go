package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/Shopify/themekit/kit"
)

const (
	banner                 string = "----------------------------------------"
	updateAvailableMessage string = `| An update for Theme Kit is available |
|                                      |
| To apply the update simply type      |
| the following command:               |
|                                      |
| theme update                         |`
)

var (
	environments kit.Environments
	themeClients []kit.ThemeClient
	directory    string
	configPath   string
	environment  string
	allenvs      bool
	notifyFile   string
	password     string
	themeid      string
	domain       string
	bucketsize   int
	refillrate   int
	concurrency  int
	proxy        string
	timeout      int

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

	ThemeCmd.PersistentFlags().StringVarP(&directory, "dir", "d", pwd, "directory that command will take effect.")
	ThemeCmd.PersistentFlags().StringVarP(&configPath, "config", "c", configPath, "path to config.yml")
	ThemeCmd.PersistentFlags().StringVarP(&environment, "env", "e", kit.DefaultEnvironment, "envionment to run the command")

	ThemeCmd.PersistentFlags().StringVar(&password, "password", "", "theme password. This will override what is in your config.yml")
	ThemeCmd.PersistentFlags().StringVar(&themeid, "themeid", "", "theme id. This will override what is in your config.yml")
	ThemeCmd.PersistentFlags().StringVar(&domain, "domain", "", "your shopify domain. This will override what is in your config.yml")
	ThemeCmd.PersistentFlags().IntVar(&bucketsize, "bucket", kit.DefaultBucketSize, "the bucket size for throttling. This will override what is in your config.yml")
	ThemeCmd.PersistentFlags().IntVar(&refillrate, "refill", kit.DefaultRefillRate, "the refill rate for throttling. This will override what is in your config.yml")
	ThemeCmd.PersistentFlags().IntVar(&concurrency, "concurrency", 1, "the refill rate for throttling. This will override what is in your config.yml")
	ThemeCmd.PersistentFlags().StringVar(&proxy, "proxy", "", "proxy for all theme requests. This will override what is in your config.yml")
	ThemeCmd.PersistentFlags().IntVarP(&timeout, "timeout", "t", kit.DefaultTimeoutInt, "the timeout to kill any stalled processes. This will override what is in your config.yml")

	watchCmd.Flags().StringVarP(&notifyFile, "notify", "n", "", "file to touch when workers have gone idle")
	watchCmd.Flags().BoolVarP(&allenvs, "allenvs", "a", false, "run command with all environments")
	removeCmd.Flags().BoolVarP(&allenvs, "allenvs", "a", false, "run command with all environments")
	replaceCmd.Flags().BoolVarP(&allenvs, "allenvs", "a", false, "run command with all environments")
	uploadCmd.Flags().BoolVarP(&allenvs, "allenvs", "a", false, "run command with all environments")

	bootstrapCmd.Flags().StringVar(&bootstrapVersion, "version", latestRelease, "version of Shopify Timber to use")
	bootstrapCmd.Flags().StringVar(&bootstrapPrefix, "prefix", "", "prefix to the Timber theme being created")
	bootstrapCmd.Flags().BoolVar(&setThemeID, "setid", true, "update config with ID of created Theme")

	ThemeCmd.AddCommand(
		bootstrapCmd,
		removeCmd,
		replaceCmd,
		uploadCmd,
		watchCmd,
		downloadCmd,
		versionCmd,
		updateCmd,
		configureCmd,
	)
}

func initializeConfig(cmdName string, timesout bool) error {
	if cmdName != "update" && isNewReleaseAvailable() {
		fmt.Println(kit.YellowText(fmt.Sprintf("%s\n%s\n%s", banner, updateAvailableMessage, banner)))
	}

	var err error
	if environments, err = kit.LoadEnvironmentsFromFile(configPath); err != nil {
		return err
	}

	eventLog := make(chan kit.ThemeEvent)
	themeClients = []kit.ThemeClient{}

	if allenvs {
		for env := range environments {
			themeClients = append(themeClients, loadThemeClient(env, eventLog))
		}
	} else {
		themeClients = []kit.ThemeClient{loadThemeClient(environment, eventLog)}
	}

	go consumeEventLog(eventLog, timesout, themeClients[0].GetConfiguration().Timeout)

	return nil
}

func loadThemeClient(env string, eventLog chan kit.ThemeEvent) kit.ThemeClient {
	client, err := loadThemeClientWithRetry(env, eventLog, false)
	if err != nil {
		if strings.Contains(err.Error(), "YAML error") {
			err = fmt.Errorf("configuration error: does your configuration properly escape wildcards? \n\t\t\t%s", err)
		} else if strings.Contains(err.Error(), "no such file or directory") {
			err = fmt.Errorf("configuration error: %s", err)
		}
		kit.Fatal(err)
	}
	return client
}

func loadThemeClientWithRetry(env string, eventLog chan kit.ThemeEvent, isRetry bool) (kit.ThemeClient, error) {
	config, err := environments.GetConfiguration(env)
	if err != nil && len(environments) == 0 && !isRetry {
		upgradeMessage := fmt.Sprintf("Looks like your configuration file is out of date. Upgrading to default environment '%s'", kit.DefaultEnvironment)
		fmt.Println(kit.YellowText(upgradeMessage))
		confirmationfn, savefn := prepareConfigurationMigration(directory)
		if confirmationfn() && savefn() == nil {
			return loadThemeClientWithRetry(env, eventLog, true)
		}
		return kit.ThemeClient{}, errors.New("loadThemeClientWithRetry: could not load or migrate the configuration")
	} else if err != nil {
		return kit.ThemeClient{}, err
	}

	if len(config.AccessToken) > 0 {
		fmt.Println("DEPRECATION WARNING: 'access_token' (in conf.yml) will soon be deprecated. Use 'password' instead, with the same Password value obtained from https://<your-subdomain>.myshopify.com/admin/apps/private/<app_id>")
	}
	return kit.NewThemeClient(eventLog, config), nil
}

func consumeEventLog(eventLog chan kit.ThemeEvent, timesout bool, timeout time.Duration) {
	eventTicked := true
	for {
		select {
		case event := <-eventLog:
			eventTicked = true
			fmt.Printf("%s\n", event)
		case <-time.Tick(timeout):
			if !timesout {
				break
			}
			if !eventTicked {
				fmt.Printf("Theme Kit timed out after %v seconds\n", timeout)
				os.Exit(1)
			}
			eventTicked = false
		}
	}
}
