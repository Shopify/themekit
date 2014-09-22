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
	bucket := phoenix.NewLeakyBucket(config.BucketSize, config.RefillRate, 1)
	bucket.TopUp()
	foreman := phoenix.NewForeman(bucket)
	client := phoenix.NewThemeClient(config)
	watcher := NewFileWatcher(dir, true)
	foreman.JobQueue = watcher
	foreman.IssueWork()

	go func() {
		asset := <-foreman.WorkerQueue
		message := fmt.Sprintf("Recieved %s event on %s", asset.Type(), asset.Asset().Key)
		fmt.Println(message)
		client.Perform(asset)
	}()

	<-done
}
