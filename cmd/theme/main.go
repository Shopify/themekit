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

// ArgsParser maps functions that take a command name and string args, to a prepared Args struct and set of CLI flags
type ArgsParser func(string, []string) commands.Args

// Command maps command string names to Commands that return a done channel, that is closed when the command operations are complete
type Command func(commands.Args, chan bool)

// CommandDefinition ...
type CommandDefinition struct {
	ArgsParser
	Command
	PermitsZeroArgs bool
	TimesOut        bool
}

var commandDefinitions = map[string]CommandDefinition{
	"upload": CommandDefinition{
		ArgsParser:      fileManipulationArgsParser,
		Command:         commands.UploadCommand,
		PermitsZeroArgs: true,
		TimesOut:        true,
	},
	"download": CommandDefinition{
		ArgsParser:      fileManipulationArgsParser,
		Command:         commands.DownloadCommand,
		PermitsZeroArgs: true,
		TimesOut:        true,
	},
	"remove": CommandDefinition{
		ArgsParser:      fileManipulationArgsParser,
		Command:         commands.RemoveCommand,
		PermitsZeroArgs: false,
		TimesOut:        true,
	},
	"replace": CommandDefinition{
		ArgsParser:      fileManipulationArgsParser,
		Command:         commands.ReplaceCommand,
		PermitsZeroArgs: true,
		TimesOut:        true,
	},
	"watch": CommandDefinition{
		ArgsParser:      watchArgsParser,
		Command:         commands.WatchCommand,
		PermitsZeroArgs: true,
		TimesOut:        false,
	},
	"configure": CommandDefinition{
		ArgsParser:      configurationArgsParser,
		Command:         commands.ConfigureCommand,
		PermitsZeroArgs: false,
		TimesOut:        false,
	},
	"bootstrap": CommandDefinition{
		ArgsParser:      bootstrapParser,
		Command:         commands.BootstrapCommand,
		PermitsZeroArgs: false,
		TimesOut:        false,
	},
	"version": CommandDefinition{
		ArgsParser:      noOpParser,
		Command:         commands.VersionCommand,
		PermitsZeroArgs: true,
		TimesOut:        false,
	},
	"update": CommandDefinition{
		ArgsParser:      noOpParser,
		Command:         commands.UpdateCommand,
		PermitsZeroArgs: true,
		TimesOut:        false,
	},
}

func main() {
	setupGlobalEventLog()
	setupErrorReporter()

	command, rest := setupAndParseArgs(os.Args[1:])
	verifyCommand(command, rest)

	if command != "update" {
		go checkForUpdate()
	}

	commandDefinition := commandDefinitions[command]
	args := commandDefinition.ArgsParser(command, rest)
	args.EventLog = globalEventLog

	done := make(chan bool)
	go func() {
		commandDefinition.Command(args, done)
	}()

	output := bufio.NewWriter(os.Stdout)
	timeout := themekit.DefaultTimeout

	if args.ThemeClient.GetConfiguration().Timeout != 0*time.Second {
		timeout = args.ThemeClient.GetConfiguration().Timeout
	}

	go consumeEventLog(output, commandDefinition, timeout, done)

	<-done
	time.Sleep(50 * time.Millisecond)
	output.Flush()
}

var eventTicked bool

func consumeEventLog(output *bufio.Writer, commandDef CommandDefinition, timeout time.Duration, done chan bool) {
	for {
		select {
		case event := <-globalEventLog:
			eventTick()

			if len(event.String()) > 0 {
				output.WriteString(fmt.Sprintf("%s\n", event))
				output.Flush()
			}
		case <-time.Tick(timeout):
			if !commandDef.TimesOut {
				break
			}

			if !eventDidTick() {
				fmt.Printf("Theme Kit timed out after %v seconds\n", timeout)
				close(done)
			}

			resetEventTick()
		}
	}
}

func eventTick() {
	eventTicked = true
}

func resetEventTick() {
	eventTicked = false
}

func eventDidTick() bool {
	return eventTicked == true
}

func commandDescription() string {
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

func noOpParser(cmd string, args []string) commands.Args {
	return commands.DefaultArgs()
}

func fileManipulationArgsParser(cmd string, rawArgs []string) commands.Args {
	args := commands.DefaultArgs()
	currentDir, _ := os.Getwd()

	set := makeFlagSet(cmd)
	set.StringVar(&args.Environment, "env", themekit.DefaultEnvironment, "environment to run command")
	set.StringVar(&args.Directory, "dir", currentDir, "directory that config.yml is located")
	set.Parse(rawArgs)

	args.ThemeClient = loadThemeClient(args.Directory, args.Environment)
	args.Filenames = rawArgs[len(rawArgs)-set.NArg():]
	return args
}

func watchArgsParser(cmd string, rawArgs []string) commands.Args {
	args := commands.DefaultArgs()
	currentDir, _ := os.Getwd()
	var allEnvironments bool

	set := makeFlagSet(cmd)
	set.StringVar(&args.Environment, "env", themekit.DefaultEnvironment, "environment to run command")
	set.BoolVar(&allEnvironments, "allenvs", false, "start watchers for all environments")
	set.StringVar(&args.Directory, "dir", currentDir, "directory that config.yml is located")
	set.StringVar(&args.NotifyFile, "notify", "", "file to touch when workers have gone idle")
	set.Parse(rawArgs)

	if len(args.Environment) != 0 && allEnvironments {
		args.Environment = ""
	}

	environments, err := loadEnvironments(args.Directory)
	if allEnvironments && err == nil {
		args.Environments = environments
	}

	args.ThemeClient = loadThemeClient(args.Directory, args.Environment)

	return args
}

func configurationArgsParser(cmd string, rawArgs []string) commands.Args {
	args := commands.DefaultArgs()
	currentDir, _ := os.Getwd()

	set := makeFlagSet(cmd)
	set.StringVar(&args.Directory, "dir", currentDir, "directory to create config.yml")
	set.StringVar(&args.Environment, "env", themekit.DefaultEnvironment, "environment for this configuration")
	set.StringVar(&args.Domain, "domain", "", "your myshopify domain")
	set.StringVar(&args.ThemeID, "theme_id", "", "your theme's id (i.e. https://<your shop>.myshopify.com/admin/themes/<theme_id>/)")
	set.StringVar(&args.Password, "password", "", "password (or access token) to make successful API calls")
	set.StringVar(&args.AccessToken, "access_token", "", "access_token to make successful API calls (optional, and soon to be deprecated in favour of 'password')")
	set.IntVar(&args.BucketSize, "bucketSize", themekit.DefaultBucketSize, "leaky bucket capacity")
	set.IntVar(&args.RefillRate, "refillRate", themekit.DefaultRefillRate, "leaky bucket refill rate / second")
	set.Parse(rawArgs)

	return args
}

func bootstrapParser(cmd string, rawArgs []string) commands.Args {
	args := commands.DefaultArgs()
	currentDir, _ := os.Getwd()

	set := makeFlagSet(cmd)
	set.StringVar(&args.Directory, "dir", currentDir, "location of config.yml")
	set.BoolVar(&args.SetThemeID, "setid", true, "update config.yml with ID of created Theme")
	set.StringVar(&args.Environment, "env", themekit.DefaultEnvironment, "environment to execute command")
	set.StringVar(&args.Version, "version", commands.LatestRelease, "version of Shopify Timber to use")
	set.StringVar(&args.Prefix, "prefix", "", "prefix to the Timber theme being created")
	set.Parse(rawArgs)

	args.ThemeClient = loadThemeClient(args.Directory, args.Environment)
	return args
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

	if len(config.AccessToken) > 0 {
		fmt.Println("DEPRECATION WARNING: 'access_token' (in conf.yml) will soon be deprecated. Use 'password' instead, with the same Password value obtained from https://<your-subdomain>.myshopify.com/admin/apps/private/<app_id>")
	}

	return themekit.NewThemeClient(config), nil
}

func loadEnvironments(directory string) (themekit.Environments, error) {
	return themekit.LoadEnvironmentsFromFile(filepath.Join(directory, "config.yml"))
}

func setupAndParseArgs(args []string) (command string, rest []string) {
	if len(args) <= 0 {
		return "", []string{"--help"}
	}
	set := makeFlagSet("")
	set.Usage = func() {
		fmt.Println(commandDescription())
	}
	set.Parse(args)

	command = args[0]
	rest = args[1:]
	return
}

func commandIsInvalid(command string) bool {
	_, found := commandDefinitions[command]
	return !found
}

func cannotProcessCommandWithoutAdditionalArguments(command string, additionalArgs []string) bool {
	commandDefinition := commandDefinitions[command]
	return len(additionalArgs) <= 0 && !commandDefinition.PermitsZeroArgs
}

func verifyCommand(command string, args []string) {
	errors := []string{}

	if commandIsInvalid(command) {
		if len(command) <= 0 {
			errors = append(errors, "  An operation must be provided")
		} else {
			errors = append(errors, fmt.Sprintf("  -'%s' is not a valid command", command))
		}
	}

	if cannotProcessCommandWithoutAdditionalArguments(command, args) {
		if commandDefinition, ok := commandDefinitions[command]; ok {
			errors = append(errors, fmt.Sprintf("\t- '%s' cannot run without additional arguments", command))
			commandDefinition.ArgsParser(command, []string{"-h"})
		}
	}

	if len(errors) > 0 {
		errorMessage := fmt.Sprintf("Invalid Invocation!\n%s", strings.Join(errors, "\n"))
		fmt.Println(themekit.RedText(errorMessage))
		setupAndParseArgs([]string{"--help"})
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
