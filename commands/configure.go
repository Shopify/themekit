package commands

import (
	"errors"
	"fmt"
	"github.com/csaunders/phoenix"
	"os"
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
		phoenix.HaltAndCatchFire(errors.New("domain and access_token cannot be blank"))
	}

	return
}

func Configure(dir, domain, accessToken string, bucketSize, refillRate int) {
	fmt.Println("Not yet implemnted. Sorry :(")
	return
}
