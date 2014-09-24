package main

import (
	"flag"
	"fmt"
	"github.com/csaunders/phoenix"
	"os"
	"strings"
)

var domain, accessToken string
var bucketSize, refillRate int
var flagSet flag.FlagSet

func main() {
	setupAndParseArgs()
	verifyArguments()
}

func setupAndParseArgs() {
	set := flag.NewFlagSet("theme-configure", flag.ExitOnError)
	set.StringVar(&domain, "domain", "", "your myshopify domain")
	set.StringVar(&accessToken, "access_token", "", "accessToken (or password) to make successful API calls")
	set.IntVar(&bucketSize, "bucketSize", phoenix.DefaultBucketSize, "leaky bucket capacity")
	set.IntVar(&refillRate, "refillRate", phoenix.DefaultRefillRate, "leaky bucket refill rate / second")
	set.Parse(os.Args[1:])
}

func verifyArguments() {
	var errors = []string{}
	if len(domain) <= 0 {
		errors = append(errors, red("domain cannot be blank"))
	}
	if len(accessToken) <= 0 {
		errors = append(errors, red("access_token cannot be blank"))
	}
	if len(errors) > 0 {
		errorMessage := fmt.Sprintf("Cannot create config.yml! errors:\n%s", strings.Join(errors, "\n"))
		fmt.Println(errorMessage)
		os.Exit(1)
	}
}

func red(s string) string {
	return fmt.Sprintf("\033[31m%s\0330", s)
}
