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

var (
	logger      *log.Logger
	errorLogger *log.Logger
)

func init() {
	logger = log.New(os.Stdout, "", 0)
	errorLogger = log.New(os.Stderr, "", 0)
}

func Logf(content string, args ...interface{}) {
	logger.Printf(content, args...)
}

func Notifyf(content string, args ...interface{}) {
	logger.Println(GreenText(fmt.Sprintf(content, args...)))
}

func Warnf(content string, args ...interface{}) {
	logger.Println(YellowText(fmt.Sprintf(content, args...)))
}

func Errorf(content string, args ...interface{}) {
	errorLogger.Println(RedText(fmt.Sprintf(content, args...)))
}

func Fatalf(content string, args ...interface{}) {
	Fatal(fmt.Errorf(content, args...))
}

// Fatal will print out the library information then log fatal the error message
// passed in. This will stop the program.
func Fatal(e error) {
	errorLogger.Println(RedText(LibraryInfo()))
	errorLogger.Fatal(RedText(e))
}
