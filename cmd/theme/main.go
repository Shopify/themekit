package main

import (
	"flag"
	"fmt"
	"github.com/csaunders/phoenix"
	"github.com/csaunders/phoenix/commands"
	"os"
	"path/filepath"
	"strings"
)

const commandDefault string = "download [<file> ...]"

var permittedZeroArgCommands = map[string]bool{
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

type Command func(map[string]interface{}) (done chan bool)

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

func main() {
	command, rest := SetupAndParseArgs(os.Args[1:])
	verifyCommand(command, rest)

	args, _ := parserMapping[command](command, rest)

	operation := commandMapping[command]
	done := operation(args)
	<-done
}

func FileManipulationCommandParser(cmd string, args []string) (result map[string]interface{}, set *flag.FlagSet) {
	result = make(map[string]interface{})
	currentDir, _ := os.Getwd()
	var environment, directory string

	set = makeFlagSet(cmd)
	set.StringVar(&environment, "env", phoenix.DefaultEnvironment, "environment to run command")
	set.StringVar(&directory, "dir", currentDir, "directory that config.yml is located")
	set.Parse(args)

	result["themeClient"] = loadThemeClient(directory, environment)
	result["filenames"] = args[len(args)-set.NArg():]
	return
}

func WatchCommandParser(cmd string, args []string) (result map[string]interface{}, set *flag.FlagSet) {
	result = make(map[string]interface{})
	currentDir, _ := os.Getwd()
	var environment, directory string

	set = makeFlagSet(cmd)
	set.StringVar(&environment, "env", phoenix.DefaultEnvironment, "environment to run command")
	set.StringVar(&directory, "dir", currentDir, "directory that config.yml is located")
	set.Parse(args)

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
	set.StringVar(&environment, "env", phoenix.DefaultEnvironment, "environment for this configuration")
	set.StringVar(&domain, "domain", "", "your myshopify domain")
	set.StringVar(&accessToken, "access_token", "", "accessToken (or password) to make successful API calls")
	set.IntVar(&bucketSize, "bucketSize", phoenix.DefaultBucketSize, "leaky bucket capacity")
	set.IntVar(&refillRate, "refillRate", phoenix.DefaultRefillRate, "leaky bucket refill rate / second")
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
	set.StringVar(&environment, "env", phoenix.DefaultEnvironment, "environment to execute command")
	set.StringVar(&version, "version", commands.LatestRelease, "version of Shopify Timber to use")
	set.StringVar(&prefix, "prefix", "", "prefix to the Timber theme being created")
	set.Parse(args)

	result["version"] = version
	result["directory"] = directory
	result["environment"] = environment
	result["prefix"] = prefix
	result["setThemeId"] = setThemeId
	result["themeClient"] = loadThemeClient(directory, environment)
	return
}

func loadThemeClient(directory, env string) phoenix.ThemeClient {
	client := loadThemeClientWithRetry(directory, env, false)
	return client
}

func loadThemeClientWithRetry(directory, env string, isRetry bool) phoenix.ThemeClient {
	environments, err := phoenix.LoadEnvironmentsFromFile(filepath.Join(directory, "config.yml"))
	if err != nil {
		phoenix.HaltAndCatchFire(err)
	}
	config, err := environments.GetConfiguration(env)
	if err != nil && !isRetry {
		upgradeMessage := fmt.Sprintf("Looks like your configuration file is out of date. Upgrading to default environment '%s'", phoenix.DefaultEnvironment)
		fmt.Println(phoenix.YellowText(upgradeMessage))
		commands.MigrateConfiguration(directory)
		client := loadThemeClientWithRetry(directory, env, true)
		return client
	} else if err != nil {
		phoenix.HaltAndCatchFire(err)
	}

	return phoenix.NewThemeClient(config)
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
		fmt.Println(phoenix.RedText(errorMessage))
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
