package main

import (
	"flag"
	"fmt"
	"github.com/csaunders/phoenix"
	"log"
	"os"
	"strings"
)

const commandDefault string = "download"

var permittedZeroArgCommands = map[string]bool{
	"download": true,
	"replace":  true,
}

var commandDescriptionPrefix = []string{
	"An operation to be performed against the theme.",
	"  Valid commands are:",
}

var permittedCommands = map[string]string{
	"upload":   "Add file(s) to theme",
	"download": "Download file(s) from theme",
	"remove":   "Remove file(s) from theme",
	"replace":  "Overwrite theme file(s)",
}

type Operation func(client phoenix.ThemeClient, filenames []string) (done chan bool)

var operations = map[string]Operation{
	"upload":   UploadOperation,
	"download": DownloadOperation,
	"remove":   RemoveOperation,
	"replace":  ReplaceOperation,
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

var command string
var filesToProcess = []string{}

func main() {
	SetupAndParseArgs(os.Args[1:])
	verifyArguments()

	config, err := phoenix.LoadConfigurationFromCurrentDirectory()
	if err != nil {
		log.Fatal(err)
	}

	client := phoenix.NewThemeClient(config)
	operation := operations[command]
	done := operation(client, filesToProcess)
	<-done
}

func SetupAndParseArgs(args []string) {
	set := flag.NewFlagSet("theme-manipulate", flag.ExitOnError)
	set.StringVar(&command, "command", commandDefault, CommandDescription(commandDefault))
	set.Parse(args)

	if len(args) != set.NArg() {
		filesToProcess = args[len(args)-set.NArg():]
	} else if len(args) > 0 && operations[args[0]] != nil {
		command = args[0]
		filesToProcess = args[1:]
	}
}

func CommandIsInvalid(command string) bool {
	return permittedCommands[command] == ""
}

func CannotProcessCommandWithoutFilenames(command string, files []string) bool {
	return len(filesToProcess) <= 0 && permittedZeroArgCommands[command] == false
}

func verifyArguments() {
	errors := []string{}

	if CommandIsInvalid(command) {
		errors = append(errors, fmt.Sprintf("\t-'%s' is not a valid command", command))
	}

	if CannotProcessCommandWithoutFilenames(command, filesToProcess) {
		errors = append(errors, "\t- There needs to be at least one file to process")
	}

	if len(errors) > 0 {
		errorMessage := fmt.Sprintf("Invalid Invocation!\n%s", strings.Join(errors, "\n"))
		fmt.Println(phoenix.RedText(errorMessage))
		SetupAndParseArgs([]string{"--help"})
		os.Exit(1)
	}
}
