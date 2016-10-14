package kit

import (
	"fmt"
	"log"

	"github.com/fatih/color"
)

// RedText is a func that wraps a string in red color tags and it will be
// red when printed out.
var RedText = color.New(color.FgRed).SprintFunc()

// YellowText is a func that wraps a string in yellow color tags and it will be
// yellow when printed out.
var YellowText = color.New(color.FgYellow).SprintFunc()

// BlueText is a func that wraps a string in blue color tags and it will be
// blue when printed out.
var BlueText = color.New(color.FgBlue).SprintFunc()

// GreenText is a func that wraps a string in green color tags and it will be
// green when printed out.
var GreenText = color.New(color.FgGreen).SprintFunc()

// Fatal will print out the library information then log fatal the error message
// passed in. This will stop the program.
func Fatal(e error) {
	fmt.Println(RedText(LibraryInfo()))
	log.Fatal(RedText(e))
}
