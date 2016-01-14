package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Shopify/themekit"
	"github.com/Shopify/themekit/commands"
)

const banner string = "----------------------------------------"
const updateAvailableMessage string = `| An update for Theme Kit is available |
|                                      |
| To apply the update simply type      |
| the following command:               |
|                                      |
| theme update                         |`

var globalEventLog chan themekit.ThemeEvent

var permittedZeroArgCommands = map[string]bool{
	"upload":   true,
	"download": true,
	"replace":  true,
	"watch":    true,
	"version":  true,
	"update":   true,
}

var commandDescriptionPrefix = []string{
	"Usage: theme <operation> [<additional arguments> ...]",
	"  Valid operations are:",
}

var permittedCommands = map[string]string{
	"upload <file> [<file2> ...]": "Add file(s) to theme",
	"download [<file> ...]":       "Download file(s) from theme",
	"remove <file> [<file2> ...]": "Remove file(s) from theme",
	"replace [<file> ...]":        "Overwrite theme file(s)",
	"watch":                       "Watch directory for changes and update remote theme",
	"configure":                   "Create a configuration file",
	"bootstrap":                   "Bootstrap a new theme using Shopify Timber",
	"version":                     "Display themekit version",
	"update":                      "Update application",
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
	"version":   NoOpParser,
	"update":    NoOpParser,
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
	"version":   commands.VersionCommand,
	"update":    commands.UpdateCommand,
}

func CommandDescription() string {
	commandDescription := make([]string, len(commandDescriptionPrefix)+len(permittedCommands))
	pos := 0
	for i := range commandDescriptionPrefix {
		commandDescription[pos] = commandDescriptionPrefix[i]
		pos++
	}

	for cmd, desc := range permittedCommands {
		commandDescription[pos] = fmt.Sprintf("    %s:\n        %s", cmd, desc)
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

func checkForUpdate() {
	if commands.IsNewReleaseAvailable() {
		message := fmt.Sprintf("%s\n%s\n%s", banner, updateAvailableMessage, banner)
		fmt.Println(themekit.YellowText(message))
	}
}

func main() {
	setupGlobalEventLog()
	setupErrorReporter()
	command, rest := SetupAndParseArgs(os.Args[1:])
	verifyCommand(command, rest)
	if command != "update" {
		go checkForUpdate()
	}

	args, _ := parserMapping[command](command, rest)
	args["eventLog"] = globalEventLog

	operation := commandMapping[command]
	done := operation(args)
	output := bufio.NewWriter(os.Stdout)
	go func() {
		ticked := false
		for {
			select {
			case event := <-globalEventLog:
				if len(event.String()) > 0 {
					output.WriteString(fmt.Sprintf("%s\n", event))
					output.Flush()
				}
			case <-time.Tick(1000 * time.Millisecond):
				if !ticked {
					ticked = true
					done <- true
				}
			}
		}
	}()
	<-done
	<-done
	output.Flush()
}

func NoOpParser(cmd string, args []string) (result map[string]interface{}, set *flag.FlagSet) {
	return make(map[string]interface{}), nil
}

func FileManipulationCommandParser(cmd string, args []string) (result map[string]interface{}, set *flag.FlagSet) {
	result = make(map[string]interface{})
	currentDir, _ := os.Getwd()
	var environment, directory string

	set = makeFlagSet(cmd)
	set.StringVar(&environment, "env", themekit.DefaultEnvironment, "environment to run command")
	set.StringVar(&directory, "dir", currentDir, "directory that config.yml is located")
	set.Parse(args)

	result["themeClient"] = loadThemeClient(directory, environment)
	result["filenames"] = args[len(args)-set.NArg():]
	return
}

func WatchCommandParser(cmd string, args []string) (result map[string]interface{}, set *flag.FlagSet) {
	result = make(map[string]interface{})
	currentDir, _ := os.Getwd()
	var allEnvironments bool
	var environment, directory, notifyFile string
	var environments themekit.Environments

	set = makeFlagSet(cmd)
	set.StringVar(&environment, "env", themekit.DefaultEnvironment, "environment to run command")
	set.BoolVar(&allEnvironments, "allenvs", false, "start watchers for all environments")
	set.StringVar(&directory, "dir", currentDir, "directory that config.yml is located")
	set.StringVar(&notifyFile, "notify", "", "file to touch when workers have gone idle")
	set.Parse(args)

	if len(environment) != 0 && allEnvironments {
		environment = ""
	}

	result["notifyFile"] = notifyFile
	environments, err := loadEnvironments(directory)
	if allEnvironments && err == nil {
		result["environments"] = environments
	}

	result["themeClient"] = loadThemeClient(directory, environment)
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
	var setThemeID bool

	set = makeFlagSet(cmd)
	set.StringVar(&directory, "dir", currentDir, "location of config.yml")
	set.BoolVar(&setThemeID, "setid", true, "update config.yml with ID of created Theme")
	set.StringVar(&environment, "env", themekit.DefaultEnvironment, "environment to execute command")
	set.StringVar(&version, "version", commands.LatestRelease, "version of Shopify Timber to use")
	set.StringVar(&prefix, "prefix", "", "prefix to the Timber theme being created")
	set.Parse(args)

	result["version"] = version
	result["directory"] = directory
	result["environment"] = environment
	result["prefix"] = prefix
	result["setThemeId"] = setThemeID
	result["themeClient"] = loadThemeClient(directory, environment)
	return
}

func loadThemeClient(directory, env string) themekit.ThemeClient {
	client, err := loadThemeClientWithRetry(directory, env, false)
	handleError(err)
	return client
}

func loadThemeClientWithRetry(directory, env string, isRetry bool) (themekit.ThemeClient, error) {
	environments, err := loadEnvironments(directory)
	if err != nil {
		return themekit.ThemeClient{}, err
	}
	config, err := environments.GetConfiguration(env)
	if err != nil && len(environments) > 0 {
		invalidEnvMsg := fmt.Sprintf("'%s' is not a valid environment. The following environments are available within config.yml:", env)
		fmt.Println(themekit.RedText(invalidEnvMsg))
		for e := range environments {
			fmt.Println(themekit.RedText(fmt.Sprintf(" - %s", e)))
		}
		os.Exit(1)
	} else if err != nil && !isRetry {
		upgradeMessage := fmt.Sprintf("Looks like your configuration file is out of date. Upgrading to default environment '%s'", themekit.DefaultEnvironment)
		fmt.Println(themekit.YellowText(upgradeMessage))
		confirmationfn, savefn := commands.PrepareConfigurationMigration(directory)

		if confirmationfn() && savefn() == nil {
			return loadThemeClientWithRetry(directory, env, true)
		}

		return themekit.ThemeClient{}, errors.New("loadThemeClientWithRetry: could not load or migrate the configuration")
	} else if err != nil {
		return themekit.ThemeClient{}, err
	}

	return themekit.NewThemeClient(config), nil
}

func loadEnvironments(directory string) (themekit.Environments, error) {
	return themekit.LoadEnvironmentsFromFile(filepath.Join(directory, "config.yml"))
}

func SetupAndParseArgs(args []string) (command string, rest []string) {
	if len(args) <= 0 {
		return "", []string{"--help"}
	}
	set := makeFlagSet("")
	set.Usage = func() {
		fmt.Println(CommandDescription())
	}
	set.Parse(args)

	command = args[0]
	rest = args[1:]
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
		if len(command) <= 0 {
			errors = append(errors, "  An operation must be provided")
		} else {
			errors = append(errors, fmt.Sprintf("  -'%s' is not a valid command", command))
		}
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

func handleError(err error) {
	if err == nil {
		return
	}

	if strings.Contains(err.Error(), "YAML error") {
		err = fmt.Errorf("configuration error: does your configuration properly escape wildcards? \n\t\t\t%s", err)
	} else if strings.Contains(err.Error(), "no such file or directory") {
		err = fmt.Errorf("configuration error: %s", err)
	}
	themekit.NotifyErrorImmediately(err)
}
