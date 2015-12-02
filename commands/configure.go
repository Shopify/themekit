package commands

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/Shopify/themekit"
)

type ConfigurationOptions struct {
	BasicOptions
	Directory   string
	Environment string
	Domain      string
	AccessToken string
	BucketSize  int
	RefillRate  int
}

func (co ConfigurationOptions) areInvalid() bool {
	return co.Domain == "" || co.AccessToken == ""
}

func (co ConfigurationOptions) defaultConfigurationOptions() themekit.Configuration {
	return themekit.Configuration{
		Domain:      co.Domain,
		AccessToken: co.AccessToken,
		BucketSize:  co.BucketSize,
		RefillRate:  co.RefillRate,
	}
}

func (co ConfigurationOptions) configurationErrors() error {
	var errs = []string{}
	if len(co.Domain) <= 0 {
		errs = append(errs, "\t-domain cannot be blank")
	}
	if len(co.AccessToken) <= 0 {
		errs = append(errs, "\t-access_token cannot be blank")
	}
	if len(errs) > 0 {
		fullPath := filepath.Join(co.Directory, "config.yml")
		return fmt.Errorf("Cannot create %s!\nErrors:\n%s", fullPath, strings.Join(errs, "\n"))
	}
	return nil
}

func defaultOptions() ConfigurationOptions {
	currentDir, _ := os.Getwd()
	return ConfigurationOptions{
		Domain:      "",
		AccessToken: "",
		Directory:   currentDir,
		Environment: themekit.DefaultEnvironment,
		BucketSize:  themekit.DefaultBucketSize,
		RefillRate:  themekit.DefaultRefillRate,
	}
}

func ConfigureCommand(args map[string]interface{}) chan bool {
	options := defaultOptions()

	extractString(&options.Environment, "environment", args)
	extractString(&options.Directory, "directory", args)
	extractString(&options.Domain, "domain", args)
	extractString(&options.AccessToken, "access_token", args)
	extractInt(&options.BucketSize, "bucket_size", args)
	extractInt(&options.RefillRate, "refill_rate", args)
	extractEventLog(&options.EventLog, args)

	if options.areInvalid() {
		themekit.NotifyError(options.configurationErrors())
	}

	Configure(options)
	done := make(chan bool)
	close(done)
	return done
}

func Configure(options ConfigurationOptions) {
	config := options.defaultConfigurationOptions()
	AddConfiguration(options.Directory, options.Environment, config)
}

func AddConfiguration(dir, environment string, config themekit.Configuration) {
	environmentLocation := filepath.Join(dir, "config.yml")
	env, err := loadOrInitializeEnvironment(environmentLocation)
	env.SetConfiguration(environment, config)

	err = env.Save(environmentLocation)
	if err != nil {
		themekit.NotifyError(err)
	}
}

func MigrateConfigurationCommand(args map[string]interface{}) (done chan bool, log chan themekit.ThemeEvent) {
	dir, _ := os.Getwd()
	extractString(&dir, "directory", args)

	MigrateConfiguration(dir)

	done = make(chan bool)
	log = make(chan themekit.ThemeEvent)
	close(done)
	close(log)
	return
}

func PrepareConfigurationMigration(dir string) (func() bool, func() error) {
	environmentLocation := filepath.Join(dir, "config.yml")
	env, err := loadOrInitializeEnvironment(environmentLocation)
	if err != nil {
		themekit.NotifyError(err)
		return func() bool { return false }, func() error { return err }
	}

	confirmationFn := func() bool {
		before, _ := ioutil.ReadFile(environmentLocation)
		after := env.String()
		fmt.Println(themekit.YellowText("Compare changes to configuration:"))
		fmt.Println(themekit.YellowText("Before:\n"), themekit.GreenText(string(before)))
		fmt.Println(themekit.YellowText("After:\n"), themekit.RedText(after))
		reader := bufio.NewReader(os.Stdin)
		fmt.Println(themekit.YellowText("Does this look correct? (y/n)"))
		text, _ := reader.ReadString('\n')
		return strings.TrimSpace(text) == "y"
	}

	saveFn := func() error {
		return env.Save(environmentLocation)
	}
	return confirmationFn, saveFn
}

func MigrateConfiguration(dir string) error {
	environmentLocation := filepath.Join(dir, "config.yml")
	env, err := loadOrInitializeEnvironment(environmentLocation)
	if err != nil {
		themekit.NotifyError(err)
		return err
	}

	err = env.Save(environmentLocation)
	return err
}

func loadOrInitializeEnvironment(location string) (themekit.Environments, error) {
	contents, err := ioutil.ReadFile(location)
	if err != nil {
		return themekit.Environments{}, err
	}

	env, err := themekit.LoadEnvironments(contents)
	if (err != nil && canProcessWithError(err)) || len(env) <= 0 {
		conf, _ := themekit.LoadConfiguration(contents)
		env[themekit.DefaultEnvironment] = conf
	}
	return env, err
}

func canProcessWithError(e error) bool {
	return strings.Contains(e.Error(), "YAML error") == false
}
