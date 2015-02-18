package commands

import (
	"errors"
	"fmt"
	"github.com/csaunders/phoenix"
	"os"
)

func WatchCommand(args map[string]interface{}) chan bool {
	var ok bool
	var client phoenix.ThemeClient
	var dir string

	if client, ok = args["themeClient"].(phoenix.ThemeClient); !ok {
		phoenix.HaltAndCatchFire(errors.New("themeClient is not of valid type"))
	}

	if args["directory"] == nil {
		dir, _ = os.Getwd()
	} else if dir, ok = args["directory"].(string); !ok {
		phoenix.HaltAndCatchFire(errors.New("directory is not of valid type"))
	}

	return Watch(client, dir)
}

func Watch(client phoenix.ThemeClient, dir string) (done chan bool) {
	done = make(chan bool)
	config := client.GetConfiguration()

	bucket := phoenix.NewLeakyBucket(config.BucketSize, config.RefillRate, 1)
	bucket.TopUp()
	foreman := phoenix.NewForeman(bucket)
	watcher := constructFileWatcher(dir, config)
	foreman.JobQueue = watcher
	foreman.IssueWork()

	fmt.Println("Waiting for local changes")
	for i := 0; i < config.Concurrency; i++ {
		go spawnWorker(i, foreman.WorkerQueue, client)
	}

	return
}

func spawnWorker(workerId int, queue chan phoenix.AssetEvent, client phoenix.ThemeClient) {
	fmt.Println(fmt.Sprintf("~~~~ Spawning Worker %d ~~~~", workerId))
	for {
		asset := <-queue
		if asset.Asset().IsValid() {
			message := fmt.Sprintf("Received %s event on '%s'", asset.Type(), asset.Asset().Key)
			fmt.Println(message)
			fmt.Println(client.Perform(asset))
		}
	}
}

func constructFileWatcher(dir string, config phoenix.Configuration) chan phoenix.AssetEvent {
	filter := phoenix.NewEventFilterFromPatternsAndFiles(config.IgnoredFiles, config.Ignores)
	return phoenix.NewFileWatcher(dir, true, filter)
}
