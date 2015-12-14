// +build !windows
package main

import (
	"fmt"
	"syscall"
)

func init() {
	var rLimit syscall.Rlimit
	if err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &rLimit); err != nil {
		fmt.Printf("Could not read max file limit: %s\n", err)
	}
	rLimit.Cur = 2048
	if err := syscall.Setrlimit(syscall.RLIMIT_NOFILE, &rLimit); err != nil {
		fmt.Printf("[warning] %s\n", err)
		fmt.Printf("[warning] could not set file descriptor limits. Themekit will work, but you might encounter issues if your project holds many files. You can set the limits manually using ulimit -n 2048.\n")
	}

}
