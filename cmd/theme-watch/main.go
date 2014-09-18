package main

import (
	"fmt"
	"github.com/csaunders/phoenix"
	"os"
)

func main() {
	done := make(chan bool)
	dir, _ := os.Getwd()
	config, _ := phoenix.LoadConfigurationFromFile(fmt.Sprintf("%s/config.yml", dir))
	fmt.Println(config.String())
	bucket := phoenix.NewLeakyBucket(config.BucketSize, config.RefillRate, 1)
	bucket.TopUp()
	foreman := phoenix.NewForeman(bucket)
	client := phoenix.NewThemeClient(config)
	watcher := NewFileWatcher(dir, true)
	foreman.JobQueue = watcher
	foreman.IssueWork()

	go client.Process(foreman.WorkerQueue)

	<-done
}
