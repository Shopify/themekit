package main

import (
	"errors"
	"flag"
	"fmt"
	"github.com/csaunders/themekit"
	"github.com/csaunders/themekit/commands"
	"os"
	"path/filepath"
	"strings"
)

const commandDefault string = "download [<file> ...]"

var globalEventLog chan themekit.ThemeEvent

var permittedZeroArgCommands = map[string]bool{
	"upload":   true,
	"download": true,
	"replace":  true,
	"watch":    true,
}

var commandDescriptionPrefix = []string{
	"An operation to be performed against the theme.",
	"  Valid commands are:",
}

var permittedCommands = map[string]string{
	"upload <file> [<file2> ...]": "Add file(s) to theme",
	"download [<file> ...]":       "Download file(s) from theme",
	"remove <file> [<file2> ...]": "Remove file(s) from theme",
	"replace [<file> ...]":        "Overwrite theme file(s)",
	"watch":                       "Watch directory for changes and update remote theme",
	"configure":                   "Create a configuration file",
	"bootstrap":                   "Bootstrap a new theme using Shopify Timber",
}

type CommandParser func(string, []string) (map[string]interface{}, *flag.FlagSet)

var parserMapping = map[string]CommandParser{
	"upload":    FileManipulationCommandParser,
	"download":  FileManipulationCommandParser,
	"remove":    FileManipulationCommandParser,
	"replace":   FileManipulationCommandParser,
	"watch":     WatchCommandParser,
	"configure": ConfigurationCommandParser,
	"bootstrap": BootstrapParser,
}

type Command func(map[string]interface{}) chan bool

var commandMapping = map[string]Command{
	"upload":    commands.UploadCommand,
	"download":  commands.DownloadCommand,
	"remove":    commands.RemoveCommand,
	"replace":   commands.ReplaceCommand,
	"watch":     commands.WatchCommand,
	"configure": commands.ConfigureCommand,
	"bootstrap": commands.BootstrapCommand,
}

func CommandDescription(defaultCommand string) string {
	commandDescription := make([]string, len(commandDescriptionPrefix)+len(permittedCommands))
	pos := 0
	for i := range commandDescriptionPrefix {
		commandDescription[pos] = commandDescriptionPrefix[i]
		pos++
	}

	for cmd, desc := range permittedCommands {
		def := ""
		if cmd == defaultCommand {
			def = " [default]"
		}
		commandDescription[pos] = fmt.Sprintf("    %s:\n        %s%s", cmd, desc, def)
		pos++
	}

	return strings.Join(commandDescription, "\n")
}

func setupErrorReporter() {
	themekit.SetErrorReporter(themekit.HaltExecutionReporter{})
}

func setupGlobalEventLog() {
	globalEventLog = make(chan themekit.ThemeEvent)
}

func main() {
	setupGlobalEventLog()
	setupErrorReporter()
	command, rest := SetupAndParseArgs(os.Args[1:])
	verifyCommand(command, rest)

	args, _ := parserMapping[command](command, rest)
	args["eventLog"] = globalEventLog

	operation := commandMapping[command]
	done := operation(args)
	go func() {
		for {
			event, more := <-globalEventLog
			if !more {
				return
			}
			fmt.Println(event)
		}
	}()
	<-done
}

func FileManipulationCommandParser(cmd string, args []string) (result map[string]interface{}, set *flag.FlagSet) {
	result = make(map[string]interface{})
	currentDir, _ := os.Getwd()
	var environment, directory string

	set = makeFlagSet(cmd)
	set.StringVar(&environment, "env", themekit.DefaultEnvironment, "environment to run command")
	set.StringVar(&directory, "dir", currentDir, "directory that config.yml is located")
	set.Parse(args)

	client, err := loadThemeClient(directory, environment)
	if err != nil {
		themekit.NotifyError(err)
		return
	}

	result["themeClient"] = client
	result["filenames"] = args[len(args)-set.NArg():]
	return
}

func WatchCommandParser(cmd string, args []string) (result map[string]interface{}, set *flag.FlagSet) {
	result = make(map[string]interface{})
	currentDir, _ := os.Getwd()
	var environment, directory string

	set = makeFlagSet(cmd)
	set.StringVar(&environment, "env", themekit.DefaultEnvironment, "environment to run command")
	set.StringVar(&directory, "dir", currentDir, "directory that config.yml is located")
	set.Parse(args)

	client, err := loadThemeClient(directory, environment)
	if err != nil {
		themekit.NotifyError(err)
		return
	}

	result["themeClient"] = client

	return
}

func ConfigurationCommandParser(cmd string, args []string) (result map[string]interface{}, set *flag.FlagSet) {
	result = make(map[string]interface{})
	currentDir, _ := os.Getwd()
	var directory, environment, domain, accessToken string
	var bucketSize, refillRate int

	set = makeFlagSet(cmd)
	set.StringVar(&directory, "dir", currentDir, "directory to create config.yml")
	set.StringVar(&environment, "env", themekit.DefaultEnvironment, "environment for this configuration")
	set.StringVar(&domain, "domain", "", "your myshopify domain")
	set.StringVar(&accessToken, "access_token", "", "accessToken (or password) to make successful API calls")
	set.IntVar(&bucketSize, "bucketSize", themekit.DefaultBucketSize, "leaky bucket capacity")
	set.IntVar(&refillRate, "refillRate", themekit.DefaultRefillRate, "leaky bucket refill rate / second")
	set.Parse(args)

	result["directory"] = directory
	result["environment"] = environment
	result["domain"] = domain
	result["access_token"] = accessToken
	result["bucket_size"] = bucketSize
	result["refill_rate"] = refillRate
	return
}

func BootstrapParser(cmd string, args []string) (result map[string]interface{}, set *flag.FlagSet) {
	result = make(map[string]interface{})
	currentDir, _ := os.Getwd()
	var version, directory, environment, prefix string
	var setThemeId bool

	set = makeFlagSet(cmd)
	set.StringVar(&directory, "dir", currentDir, "location of config.yml")
	set.BoolVar(&setThemeId, "setid", true, "update config.yml with ID of created Theme")
	set.StringVar(&environment, "env", themekit.DefaultEnvironment, "environment to execute command")
	set.StringVar(&version, "version", commands.LatestRelease, "version of Shopify Timber to use")
	set.StringVar(&prefix, "prefix", "", "prefix to the Timber theme being created")
	set.Parse(args)

	result["version"] = version
	result["directory"] = directory
	result["environment"] = environment
	result["prefix"] = prefix
	result["setThemeId"] = setThemeId

	client, err := loadThemeClient(directory, environment)
	if err != nil {
		themekit.NotifyError(err)
		return
	}
	result["themeClient"] = client
	return
}

func loadThemeClient(directory, env string) (themekit.ThemeClient, error) {
	client, err := loadThemeClientWithRetry(directory, env, false)
	if err != nil && strings.Contains(err.Error(), "YAML error") {
		err = errors.New(fmt.Sprintf("configuration error: does your configuration properly escape wildcards? \n\t\t\t%s", err))
	}
	return client, err
}

func loadThemeClientWithRetry(directory, env string, isRetry bool) (themekit.ThemeClient, error) {
	environments, err := themekit.LoadEnvironmentsFromFile(filepath.Join(directory, "config.yml"))
	if err != nil {
		return themekit.ThemeClient{}, err
	}
	config, err := environments.GetConfiguration(env)
	if err != nil && !isRetry {
		upgradeMessage := fmt.Sprintf("Looks like your configuration file is out of date. Upgrading to default environment '%s'", themekit.DefaultEnvironment)
		fmt.Println(themekit.YellowText(upgradeMessage))
		err := commands.MigrateConfiguration(directory)
		if err != nil {
			return themekit.ThemeClient{}, err
		}
		return loadThemeClientWithRetry(directory, env, true)
	} else if err != nil {
		return themekit.ThemeClient{}, err
	}

	return themekit.NewThemeClient(config), nil
}

func SetupAndParseArgs(args []string) (command string, rest []string) {

	set := makeFlagSet("")
	set.StringVar(&command, "command", "download", CommandDescription(commandDefault))
	set.Parse(args)

	if len(args) != set.NArg() {
		rest = args[len(args)-set.NArg():]
	} else if len(args) > 0 {
		command = args[0]
		rest = args[1:]
	}
	return
}

func CommandIsInvalid(command string) bool {
	return commandMapping[command] == nil
}

func CannotProcessCommandWithoutAdditionalArguments(command string, additionalArgs []string) bool {
	return len(additionalArgs) <= 0 && permittedZeroArgCommands[command] == false
}

func verifyCommand(command string, args []string) {
	errors := []string{}

	if CommandIsInvalid(command) {
		errors = append(errors, fmt.Sprintf("\t-'%s' is not a valid command", command))
	}

	if CannotProcessCommandWithoutAdditionalArguments(command, args) {
		if parser, ok := parserMapping[command]; ok {
			errors = append(errors, fmt.Sprintf("\t- '%s' cannot run without additional arguments", command))
			parser(command, []string{"-h"})
		}
	}

	if len(errors) > 0 {
		errorMessage := fmt.Sprintf("Invalid Invocation!\n%s", strings.Join(errors, "\n"))
		fmt.Println(themekit.RedText(errorMessage))
		SetupAndParseArgs([]string{"--help"})
		os.Exit(1)
	}
}

func makeFlagSet(command string) *flag.FlagSet {
	if command != "" {
		command = " " + command
	}
	return flag.NewFlagSet(fmt.Sprintf("theme%s", command), flag.ExitOnError)
}
