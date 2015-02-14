package main

import (
	"flag"
	"fmt"
	"github.com/csaunders/phoenix"
	"github.com/csaunders/phoenix/commands"
	"os"
	"strings"
)

var domain, accessToken string
var bucketSize, refillRate int

func main() {
	setupAndParseArgs(os.Args[1:])
	verifyArguments()

	commands.Configure("", domain, accessToken, bucketSize, refillRate)
}

func setupAndParseArgs(args []string) {
	set := flag.NewFlagSet("theme-configure", flag.ExitOnError)
	set.StringVar(&domain, "domain", "", "your myshopify domain")
	set.StringVar(&accessToken, "access_token", "", "accessToken (or password) to make successful API calls")
	set.IntVar(&bucketSize, "bucketSize", phoenix.DefaultBucketSize, "leaky bucket capacity")
	set.IntVar(&refillRate, "refillRate", phoenix.DefaultRefillRate, "leaky bucket refill rate / second")
	set.Parse(args)
}

func verifyArguments() {
	var errors = []string{}
	if len(domain) <= 0 {
		errors = append(errors, "\t-domain cannot be blank")
	}
	if len(accessToken) <= 0 {
		errors = append(errors, "\t-access_token cannot be blank")
	}
	if len(errors) > 0 {
		errorMessage := red(fmt.Sprintf("Cannot create config.yml!\nErrors:\n%s", strings.Join(errors, "\n")))
		fmt.Println(errorMessage)
		setupAndParseArgs([]string{"--help"})
		os.Exit(1)
	}
}

// TODO: Deprecate this and use RedText from utils.go instead
func red(s string) string {
	return fmt.Sprintf("\033[31m%s\033[0m", s)
}
