package kit

import (
	"fmt"
	"log"
	"time"

	"github.com/fatih/color"
	"github.com/mattn/go-colorable"
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

// CyanText is a func that wraps a string in cyan color tags and it will be
// cyan when printed out.
var CyanText = color.New(color.FgCyan).SprintFunc()

const timestampFormat = "15:04:05"

var (
	stdOut = colorable.NewColorableStdout()
	stdErr = colorable.NewColorableStderr()
)

func init() {
	log.SetFlags(0)
	log.SetOutput(stdErr)
}

func timestamp() string {
	return CyanText(time.Now().Format(timestampFormat)) + " "
}

// Printf will output a formatted message to the output log (default stdout)
func Printf(content string, args ...interface{}) {
	fmt.Fprintln(stdOut, timestamp()+fmt.Sprintf(content, args...))
}

// Print will output a message to the output log (default stdout)
func Print(args ...interface{}) {
	fmt.Fprintln(stdOut, timestamp()+fmt.Sprint(args...))
}

// LogErrorf will output a formatted red message to the output log (default stdout)
func LogErrorf(content string, args ...interface{}) {
	log.Println(timestamp() + fmt.Sprintf(content, args...))
}

// LogError will output a red message to the output log (default stdout)
func LogError(args ...interface{}) {
	log.Println(timestamp() + fmt.Sprint(args...))
}

// LogFatalf will output a formatted red message to the output log along with the
// library information. Then it will quit the application.
func LogFatalf(content string, args ...interface{}) {
	log.Fatal(fmt.Sprintf(content, args...))
}

// LogFatal will output a red message to the output log along with the
// library information. Then it will quit the application.
func LogFatal(args ...interface{}) {
	log.Fatal(fmt.Sprint(args...))
}
