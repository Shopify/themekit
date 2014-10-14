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
	"upload": UploadOperation,
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
	setupAndParseArgs(os.Args[1:])
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

func setupAndParseArgs(args []string) {
	set := flag.NewFlagSet("theme-manipulate", flag.ExitOnError)
	set.StringVar(&command, "command", commandDefault, CommandDescription(commandDefault))
	set.Parse(args)
	filesToProcess = args[len(args)-set.NArg():]
}

func verifyArguments() {
	errors := []string{}

	if permittedCommands[command] == "" {
		errors = append(errors, fmt.Sprintf("\t-'%s' is not a valid command", command))
	}

	if len(filesToProcess) <= 0 {
		errors = append(errors, "\t- There needs to be at least one file to process")
	}

	if len(errors) > 0 {
		errorMessage := fmt.Sprintf("Invalid Invocation!\n%s", strings.Join(errors, "\n"))
		fmt.Println(phoenix.RedText(errorMessage))
		setupAndParseArgs([]string{"--help"})
		os.Exit(1)
	}
}

func loadAsset(filename string) (asset phoenix.Asset, err error) {
	root, err := os.Getwd()
	if err != nil {
		return
	}

	path := fmt.Sprintf("%s/%s", root, filename)
	file, err := os.Open(path)
	info, err := os.Stat(path)
	if err != nil {
		return
	}

	buffer := make([]byte, info.Size())
	_, err = file.Read(buffer)
	if err != nil {
		return
	}
	asset = phoenix.Asset{Value: string(buffer), Key: filename}
	return
}
