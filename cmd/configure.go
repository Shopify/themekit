package cmd

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/Shopify/themekit/kit"
)

var configureCmd = &cobra.Command{
	Use:   "configure",
	Short: "Create a configuration file",
	Long: `Configure will create a new configuration file to
access shopify using the theme kit.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		var errs = []string{}
		if len(domain) <= 0 {
			errs = append(errs, "\t-domain cannot be blank")
		}
		if len(password) <= 0 {
			errs = append(errs, "\t-password or access_token cannot be blank")
		}
		if len(errs) > 0 {
			fullPath := filepath.Join(directory, "config.yml")
			return fmt.Errorf("Cannot create %s!\nErrors:\n%s", fullPath, strings.Join(errs, "\n"))
		}

		config := kit.Configuration{
			Domain:      domain,
			AccessToken: password,
			Password:    password,
			BucketSize:  bucketsize,
			RefillRate:  refillrate,
			Timeout:     time.Duration(timeout) * time.Second,
			ThemeID:     themeid,
		}

		_, err := config.Initialize()
		if err != nil {
			return err
		}

		return addConfiguration(config)
	},
}

func addConfiguration(config kit.Configuration) error {
	env, err := loadOrInitializeEnvironment(configPath)
	if err != nil {
		return err
	}
	env.SetConfiguration(environment, config)
	return env.Save(configPath)
}

func prepareConfigurationMigration(dir string) (func() bool, func() error) {
	environmentLocation := filepath.Join(dir, "config.yml")
	env, err := loadOrInitializeEnvironment(environmentLocation)
	if err != nil {
		kit.Fatal(err)
		return func() bool { return false }, func() error { return err }
	}

	confirmationFn := func() bool {
		before, _ := ioutil.ReadFile(environmentLocation)
		after := env.String()
		fmt.Println(kit.YellowText("Compare changes to configuration:"))
		fmt.Println(kit.YellowText("Before:\n"), kit.GreenText(string(before)))
		fmt.Println(kit.YellowText("After:\n"), kit.RedText(after))
		reader := bufio.NewReader(os.Stdin)
		fmt.Println(kit.YellowText("Does this look correct? (y/n)"))
		text, _ := reader.ReadString('\n')
		return strings.TrimSpace(text) == "y"
	}

	saveFn := func() error {
		return env.Save(environmentLocation)
	}
	return confirmationFn, saveFn
}

func loadOrInitializeEnvironment(location string) (kit.Environments, error) {
	contents, err := ioutil.ReadFile(location)
	if err != nil {
		return kit.Environments{}, err
	}

	env, err := kit.LoadEnvironments(contents)

	if err != nil && !canProcessWithError(err) {
		return env, err
	}

	if err != nil || len(env) <= 0 {
		conf, _ := kit.LoadConfiguration(contents)
		env[kit.DefaultEnvironment] = conf
	}
	return env, err
}

func canProcessWithError(e error) bool {
	if strings.Contains(e.Error(), "YAML error") == false {
		return false
	}

	return true
}
