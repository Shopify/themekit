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
	watcher := constructFileWatcher(dir, config)
	foreman.JobQueue = watcher
	foreman.IssueWork()

	fmt.Println("Waiting for local changes")
	for i := 0; i < config.Concurrency; i++ {
		go spawnWorker(i, foreman.WorkerQueue, client)
	}

	<-done
}

func spawnWorker(workerId int, queue chan phoenix.AssetEvent, client phoenix.ThemeClient) {
	fmt.Println(fmt.Sprintf("~~~~ Spawning Worker %d ~~~~", workerId))
	for {
		asset := <-queue
		message := fmt.Sprintf("Recieved %s event on '%s'", asset.Type(), asset.Asset().Key)
		fmt.Println(message)
		client.Perform(asset)
	}
}

func constructFileWatcher(dir string, config phoenix.Configuration) chan phoenix.AssetEvent {
	filter := phoenix.NewEventFilterFromFilesCSV(config.Ignores)
	return NewFileWatcher(dir, true, filter)
}
