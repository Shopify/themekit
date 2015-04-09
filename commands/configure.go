package commands

import (
	"errors"
	"fmt"
	"github.com/csaunders/themekit"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
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
		return errors.New(fmt.Sprintf("Cannot create %s!\nErrors:\n%s", fullPath, strings.Join(errs, "\n")))
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
	env := loadOrInitializeEnvironment(environmentLocation)
	env.SetConfiguration(environment, config)

	err := env.Save(environmentLocation)
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

func MigrateConfiguration(dir string) {
	environmentLocation := filepath.Join(dir, "config.yml")
	env := loadOrInitializeEnvironment(environmentLocation)
	err := env.Save(environmentLocation)
	if err != nil {
		themekit.NotifyError(err)
	}
}

func loadOrInitializeEnvironment(location string) themekit.Environments {
	contents, err := ioutil.ReadFile(location)
	if err != nil {
		return themekit.Environments{}
	}

	env, err := themekit.LoadEnvironments(contents)
	if err != nil || len(env) <= 0 {
		conf, _ := themekit.LoadConfiguration(contents)
		env[themekit.DefaultEnvironment] = conf
	}
	return env
}
