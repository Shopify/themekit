package colors

import (
	"log"

	"github.com/fatih/color"
	"github.com/mattn/go-colorable"
)

var (
	// Red is the color red
	Red = color.New(color.FgRed).SprintFunc()
	// Yellow is the color yellow
	Yellow = color.New(color.FgYellow).SprintFunc()
	// Blue is the color blue
	Blue = color.New(color.FgBlue).SprintFunc()
	// Green is the color Green
	Green = color.New(color.FgGreen).SprintFunc()
	// ColorStdOut is a wrapped std out that allows colors
	ColorStdOut = log.New(colorable.NewColorableStdout(), "", 0)
	// ColorStdErr is a wrapped std err that allows colors
	ColorStdErr = log.New(colorable.NewColorableStderr(), "", 0)
	// Cyan is the color cyan
	Cyan = color.New(color.FgCyan).SprintFunc()
)
