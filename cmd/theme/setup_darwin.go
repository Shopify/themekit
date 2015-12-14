// +build darwin,amd64

package main

import (
	"fmt"
	"math"
	"syscall"
)

const MinFileDescriptors float64 = 2048

// MacOSX sets a very low default file descriptor limit per process. This function sets file descriptor limits to a more sane value.
func init() {
	var rLimit syscall.Rlimit
	if err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &rLimit); err != nil {
		fmt.Printf("Could not read max file limit: %s\n", err)
	}
	rLimit.Cur = uint64(math.Max(MinFileDescriptors, float64(rLimit.Cur)))
	if err := syscall.Setrlimit(syscall.RLIMIT_NOFILE, &rLimit); err != nil {
		fmt.Printf("[warning] %s\n", err)
		fmt.Printf("[warning] could not set file descriptor limits. Themekit will work, but you might encounter issues if your project holds many files. You can set the limits manually using ulimit -n 2048.\n")
	}

}
