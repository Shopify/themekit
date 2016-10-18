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

func Printf(content string, args ...interface{}) {
	fmt.Println(fmt.Sprintf(content, args...))
}

func Print(args ...interface{}) {
	fmt.Println(args...)
}

func Notifyf(content string, args ...interface{}) {
	fmt.Println(GreenText(fmt.Sprintf(content, args...)))
}

func LogNotify(args ...interface{}) {
	log.Println(GreenText(fmt.Sprint(args...)))
}

func LogWarnf(content string, args ...interface{}) {
	log.Println(YellowText(fmt.Sprintf(content, args...)))
}

func LogWarn(args ...interface{}) {
	log.Println(YellowText(fmt.Sprint(args...)))
}

func LogErrorf(content string, args ...interface{}) {
	log.Println(RedText(fmt.Sprintf(content, args...)))
}

func LogError(args ...interface{}) {
	log.Println(RedText(fmt.Sprint(args...)))
}

func LogFatalf(content string, args ...interface{}) {
	log.Println(RedText(LibraryInfo()))
	log.Fatal(RedText(fmt.Sprintf(content, args...)))
}

// LogFatal will print out the library information then log fatal the error message
// passed in. This will stop the program.
func LogFatal(args ...interface{}) {
	log.Println(RedText(LibraryInfo()))
	log.Fatal(RedText(fmt.Sprint(args...)))
}
