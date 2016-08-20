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

// ConfigureCommand creates a configuration file
func ConfigureCommand(args Args, done chan bool) {
	if err := args.ConfigurationErrors(); err != nil {
		themekit.NotifyError(err)
	}

	Configure(args)
	close(done)
}

// Configure ... TODO
func Configure(args Args) {
	config := args.DefaultConfigurationOptions()
	_, err := config.Initialize()
	if err != nil {
		fmt.Println(err)
		return
	}

	AddConfiguration(args.Directory, args.Environment, config)
}

// AddConfiguration ... TODO
func AddConfiguration(dir, environment string, config themekit.Configuration) {
	environmentLocation := filepath.Join(dir, "config.yml")
	env, err := loadOrInitializeEnvironment(environmentLocation)
	env.SetConfiguration(environment, config)

	err = env.Save(environmentLocation)
	if err != nil {
		themekit.NotifyError(err)
	}
}

// MigrateConfigurationCommand ... TODO
func MigrateConfigurationCommand(args Args) (done chan bool, log chan themekit.ThemeEvent) {
	MigrateConfiguration(args.Directory)

	done = make(chan bool)
	log = make(chan themekit.ThemeEvent)
	close(done)
	close(log)
	return
}

// PrepareConfigurationMigration ... TODO
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

// MigrateConfiguration ... TODO
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

	if err != nil && !canProcessWithError(err) {
		return env, err
	}

	if err != nil || len(env) <= 0 {
		conf, _ := themekit.LoadConfiguration(contents)
		env[themekit.DefaultEnvironment] = conf
	}
	return env, err
}

func canProcessWithError(e error) bool {
	if strings.Contains(e.Error(), "YAML error") == false {
		return false
	}

	return true
}
