package cmd

import (
	"log"

	"github.com/fatih/color"
	"github.com/mattn/go-colorable"
)

var (
	red    = color.New(color.FgRed).SprintFunc()
	yellow = color.New(color.FgYellow).SprintFunc()
	blue   = color.New(color.FgBlue).SprintFunc()
	green  = color.New(color.FgGreen).SprintFunc()

	stdOut = log.New(colorable.NewColorableStdout(), "", 0)
	stdErr = log.New(colorable.NewColorableStderr(), "", 0)
)
