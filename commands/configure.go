package commands

import (
	"fmt"
	"github.com/csaunders/phoenix"
	"os"
	"path/filepath"
	"strings"
)

func ConfigureCommand(args map[string]interface{}) (done chan bool) {
	currentDir, _ := os.Getwd()
	var dir, domain, accessToken string = currentDir, "", ""
	var bucketSize, refillRate int = phoenix.DefaultBucketSize, phoenix.DefaultRefillRate

	extractString(&dir, "directory", args)
	extractString(&domain, "domain", args)
	extractString(&accessToken, "access_token", args)
	extractInt(&bucketSize, "bucket_size", args)
	extractInt(&refillRate, "refill_rate", args)

	if domain == "" || accessToken == "" {
		reportArgumentsError(dir, domain, accessToken)
	}

	Configure(dir, domain, accessToken, bucketSize, refillRate)
	done = make(chan bool)
	close(done)
	return
}

func Configure(dir, domain, accessToken string, bucketSize, refillRate int) {
	config := phoenix.Configuration{Domain: domain, AccessToken: accessToken, BucketSize: bucketSize, RefillRate: refillRate}
	err := config.Save(filepath.Join(dir, "config.yml"))
	if err != nil {
		phoenix.HaltAndCatchFire(err)
	}
}

func reportArgumentsError(directory, domain, accessToken string) {
	var errors = []string{}
	if len(domain) <= 0 {
		errors = append(errors, "\t-domain cannot be blank")
	}
	if len(accessToken) <= 0 {
		errors = append(errors, "\t-access_token cannot be blank")
	}
	if len(errors) > 0 {
		fullPath := filepath.Join(directory, "config.yml")
		errorMessage := phoenix.RedText(fmt.Sprintf("Cannot create %s!\nErrors:\n%s", fullPath, strings.Join(errors, "\n")))
		fmt.Println(errorMessage)
		os.Exit(1)
	}
}
