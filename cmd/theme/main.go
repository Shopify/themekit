package main

import (
	"flag"
	"fmt"
	"github.com/csaunders/phoenix"
	"github.com/csaunders/phoenix/commands"
	"os"
	"strings"
)

const commandDefault string = "download"

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
	"upload":    "Add file(s) to theme",
	"download":  "Download file(s) from theme",
	"remove":    "Remove file(s) from theme",
	"replace":   "Overwrite theme file(s)",
	"watch":     "Watch director for changes and update remote theme",
	"configure": "Create a configuration file",
}

type CommandParser func([]string) map[string]interface{}

var parserMapping = map[string]CommandParser{
	"upload":   FileManipulationCommandParser,
	"download": FileManipulationCommandParser,
	"remove":   FileManipulationCommandParser,
	"replace":  FileManipulationCommandParser,
}

type Command func(map[string]interface{}) (done chan bool)

var commandMapping = map[string]Command{
	"upload":    commands.UploadCommand,
	"download":  commands.DownloadCommand,
	"remove":    commands.RemoveCommand,
	"replace":   commands.ReplaceCommand,
	"watch":     commands.WatchCommand,
	"configure": commands.ConfigureCommand,
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
		commandDescription[pos] = fmt.Sprintf("    %s: %s%s", cmd, desc, def)
		pos++
	}

	return strings.Join(commandDescription, "\n")
}

func main() {
	command, rest := SetupAndParseArgs(os.Args[1:])
	verifyCommand(command, rest)

	parser := parserMapping[command]
	args := parser(rest)

	operation := commandMapping[command]
	done := operation(args)
	<-done
}

func FileManipulationCommandParser(args []string) (result map[string]interface{}) {

	result = make(map[string]interface{})
	result["themeClient"] = loadThemeClient()
	result["filenames"] = args
	return
}

func ConfigurationCommandParser(args []string) (result map[string]interface{}) {
	result = make(map[string]interface{})
	var domain, accessToken string
	var bucketSize, refillRate int

	set := flag.NewFlagSet("theme", flag.ExitOnError)
	set.StringVar(&domain, "domain", "", "your myshopify domain")
	set.StringVar(&accessToken, "access_token", "", "accessToken (or password) to make successful API calls")
	set.IntVar(&bucketSize, "bucketSize", phoenix.DefaultBucketSize, "leaky bucket capacity")
	set.IntVar(&refillRate, "refillRate", phoenix.DefaultRefillRate, "leaky bucket refill rate / second")
	set.Parse(args)

	result["domain"] = domain
	result["accessToken"] = accessToken
	result["bucketSize"] = bucketSize
	result["refillRate"] = refillRate
	return
}

func loadThemeClient() phoenix.ThemeClient {
	config, err := phoenix.LoadConfigurationFromCurrentDirectory()
	if err != nil {
		phoenix.HaltAndCatchFire(err)
	}

	return phoenix.NewThemeClient(config)
}

func SetupAndParseArgs(args []string) (command string, rest []string) {
	set := flag.NewFlagSet("theme", flag.ExitOnError)
	set.StringVar(&command, "command", commandDefault, CommandDescription(commandDefault))
	set.Parse(args)

	if len(args) != set.NArg() {
		rest = args[len(args)-set.NArg():]
	} else if len(args) > 0 && commandMapping[args[0]] != nil {
		command = args[0]
		rest = args[1:]
	}
	return
}

func CommandIsInvalid(command string) bool {
	return permittedCommands[command] == ""
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
		errors = append(errors, "\t- There needs to be at least one file to process")
	}

	if len(errors) > 0 {
		errorMessage := fmt.Sprintf("Invalid Invocation!\n%s", strings.Join(errors, "\n"))
		fmt.Println(phoenix.RedText(errorMessage))
		SetupAndParseArgs([]string{"--help"})
		os.Exit(1)
	}
}
