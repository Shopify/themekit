package commands

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/Shopify/themekit/kit"
)

// ConfigureCommand creates a configuration file
func ConfigureCommand(args Args, done chan bool) {
	if err := args.ConfigurationErrors(); err != nil {
		kit.NotifyError(err)
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
func AddConfiguration(dir, environment string, config kit.Configuration) {
	environmentLocation := filepath.Join(dir, "config.yml")
	env, err := loadOrInitializeEnvironment(environmentLocation)
	env.SetConfiguration(environment, config)

	err = env.Save(environmentLocation)
	if err != nil {
		kit.NotifyError(err)
	}
}

// MigrateConfigurationCommand ... TODO
func MigrateConfigurationCommand(args Args) (done chan bool, log chan kit.ThemeEvent) {
	MigrateConfiguration(args.Directory)

	done = make(chan bool)
	log = make(chan kit.ThemeEvent)
	close(done)
	close(log)
	return
}

// PrepareConfigurationMigration ... TODO
func PrepareConfigurationMigration(dir string) (func() bool, func() error) {
	environmentLocation := filepath.Join(dir, "config.yml")
	env, err := loadOrInitializeEnvironment(environmentLocation)
	if err != nil {
		kit.NotifyError(err)
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

// MigrateConfiguration ... TODO
func MigrateConfiguration(dir string) error {
	environmentLocation := filepath.Join(dir, "config.yml")
	env, err := loadOrInitializeEnvironment(environmentLocation)
	if err != nil {
		kit.NotifyError(err)
		return err
	}

	err = env.Save(environmentLocation)
	return err
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
