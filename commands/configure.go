package commands

import (
	"fmt"
	"github.com/csaunders/phoenix"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

func ConfigureCommand(args map[string]interface{}) (done chan bool) {
	currentDir, _ := os.Getwd()
	environment := phoenix.DefaultEnvironment
	var dir, domain, accessToken string = currentDir, "", ""
	var bucketSize, refillRate int = phoenix.DefaultBucketSize, phoenix.DefaultRefillRate

	extractString(&environment, "environment", args)
	extractString(&dir, "directory", args)
	extractString(&domain, "domain", args)
	extractString(&accessToken, "access_token", args)
	extractInt(&bucketSize, "bucket_size", args)
	extractInt(&refillRate, "refill_rate", args)

	if domain == "" || accessToken == "" {
		reportArgumentsError(dir, domain, accessToken)
	}

	Configure(dir, environment, domain, accessToken, bucketSize, refillRate)
	done = make(chan bool)
	close(done)
	return
}

func Configure(dir, environment, domain, accessToken string, bucketSize, refillRate int) {
	config := phoenix.Configuration{Domain: domain, AccessToken: accessToken, BucketSize: bucketSize, RefillRate: refillRate}
	AddConfiguration(dir, environment, config)
}

func AddConfiguration(dir, environment string, config phoenix.Configuration) {
	environmentLocation := filepath.Join(dir, "config.yml")
	env := loadOrInitializeEnvironment(environmentLocation)
	env.SetConfiguration(environment, config)

	err := env.Save(environmentLocation)
	if err != nil {
		phoenix.HaltAndCatchFire(err)
	}
}

func MigrateConfigurationCommand(args map[string]interface{}) (done chan bool) {
	dir, _ := os.Getwd()
	extractString(&dir, "directory", args)

	MigrateConfiguration(dir)

	done = make(chan bool)
	close(done)
	return
}

func MigrateConfiguration(dir string) {
	environmentLocation := filepath.Join(dir, "config.yml")
	env := loadOrInitializeEnvironment(environmentLocation)
	err := env.Save(environmentLocation)
	if err != nil {
		phoenix.HaltAndCatchFire(err)
	}
}

func loadOrInitializeEnvironment(location string) phoenix.Environments {
	contents, err := ioutil.ReadFile(location)
	if err != nil {
		return phoenix.Environments{}
	}

	env, err := phoenix.LoadEnvironments(contents)
	if err != nil || len(env) <= 0 {
		conf, _ := phoenix.LoadConfiguration(contents)
		env[phoenix.DefaultEnvironment] = conf
	}
	return env
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
