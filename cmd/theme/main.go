package main

import (
	"os"
	"runtime"
	"runtime/pprof"

	"github.com/Shopify/themekit/cmd"
	"github.com/Shopify/themekit/src/colors"
)

const (
	cpuProfileVar = "THEMEKIT_CPUPROFILE"
	memProfileVar = "THEMEKIT_MEMPROFILE"
)

var stdErr = colors.ColorStdErr

func main() {
	if CPUProfile := os.Getenv(cpuProfileVar); CPUProfile != "" {
		f, err := os.Create(CPUProfile)
		if err != nil {
			stdErr.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	if err := cmd.ThemeCmd.Execute(); err != nil {
		stdErr.Fatal(colors.Red(err.Error()))
	}

	if memProfile := os.Getenv(memProfileVar); memProfile != "" {
		f, err := os.Create(memProfile)
		if err != nil {
			stdErr.Fatal("could not create memory profile: ", err)
		}
		runtime.GC() // get up-to-date statistics
		if err := pprof.WriteHeapProfile(f); err != nil {
			stdErr.Fatal("could not write memory profile: ", err)
		}
		f.Close()
	}
}
