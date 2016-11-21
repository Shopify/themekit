package kit

import (
	"fmt"
	"log"
	"os"

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

func init() {
	log.SetFlags(0)
	log.SetOutput(os.Stderr)
}

// Printf will output a formatted message to the output log (default stdout)
func Printf(content string, args ...interface{}) {
	fmt.Println(fmt.Sprintf(content, args...))
}

// Print will output a message to the output log (default stdout)
func Print(args ...interface{}) {
	fmt.Println(args...)
}

// LogNotifyf will output a formatted green message to the output log (default stdout)
func LogNotifyf(content string, args ...interface{}) {
	fmt.Println(GreenText(fmt.Sprintf(content, args...)))
}

// LogNotify will output a green message to the output log (default stdout)
func LogNotify(args ...interface{}) {
	log.Println(GreenText(fmt.Sprint(args...)))
}

// LogWarnf will output a formatted yellow message to the output log (default stdout)
func LogWarnf(content string, args ...interface{}) {
	log.Println(YellowText(fmt.Sprintf(content, args...)))
}

// LogWarn will output a yellow message to the output log (default stdout)
func LogWarn(args ...interface{}) {
	log.Println(YellowText(fmt.Sprint(args...)))
}

// LogErrorf will output a formatted red message to the output log (default stdout)
func LogErrorf(content string, args ...interface{}) {
	log.Println(fmt.Sprintf(content, args...))
}

// LogError will output a red message to the output log (default stdout)
func LogError(args ...interface{}) {
	log.Println(fmt.Sprint(args...))
}

// LogFatalf will output a formatted red message to the output log along with the
// library information. Then it will quit the application.
func LogFatalf(content string, args ...interface{}) {
	log.Println(RedText(LibraryInfo()))
	log.Fatal(fmt.Sprintf(content, args...))
}

// LogFatal will output a red message to the output log along with the
// library information. Then it will quit the application.
func LogFatal(args ...interface{}) {
	log.Println(RedText(LibraryInfo()))
	log.Fatal(fmt.Sprint(args...))
}
