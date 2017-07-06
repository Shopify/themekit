package main

import (
	"os"
	"runtime"
	"runtime/pprof"

	"github.com/Shopify/themekit/cmd"
	"github.com/Shopify/themekit/kit"
)

const (
	cpuProfileVar = "THEMEKIT_CPUPROFILE"
	memProfileVar = "THEMEKIT_MEMPROFILE"
)

func main() {
	if CPUProfile := os.Getenv(cpuProfileVar); CPUProfile != "" {
		f, err := os.Create(CPUProfile)
		if err != nil {
			kit.LogFatal(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	if err := cmd.ThemeCmd.Execute(); err != nil {
		kit.LogFatalf("Err: %v", err)
	}

	if memProfile := os.Getenv(memProfileVar); memProfile != "" {
		f, err := os.Create(memProfile)
		if err != nil {
			kit.LogFatal("could not create memory profile: ", err)
		}
		runtime.GC() // get up-to-date statistics
		if err := pprof.WriteHeapProfile(f); err != nil {
			kit.LogFatal("could not write memory profile: ", err)
		}
		f.Close()
	}
}
