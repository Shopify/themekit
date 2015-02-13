package commands

import (
	"errors"
	"fmt"
	"github.com/csaunders/phoenix"
	"os"
)

func WatchCommand(args map[string]interface{}) chan bool {
	var ok bool
	var config phoenix.Configuration
	var dir string

	if config, ok = args["configuration"].(phoenix.Configuration); !ok {
		phoenix.HaltAndCatchFire(errors.New("configuration is not of valid type"))
	}

	if args["directory"] == nil {
		dir, _ = os.Getwd()
	} else if dir, ok = args["directory"].(string); !ok {
		phoenix.HaltAndCatchFire(errors.New("directory is not of valid type"))
	}

	return Watch(config, dir)
}

func Watch(config phoenix.Configuration, dir string) (done chan bool) {
	done = make(chan bool)

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

	return
}

func spawnWorker(workerId int, queue chan phoenix.AssetEvent, client phoenix.ThemeClient) {
	fmt.Println(fmt.Sprintf("~~~~ Spawning Worker %d ~~~~", workerId))
	for {
		asset := <-queue
		if asset.Asset().IsValid() {
			message := fmt.Sprintf("Recieved %s event on '%s'", asset.Type(), asset.Asset().Key)
			fmt.Println(message)
			fmt.Println(client.Perform(asset))
		}
	}
}

func constructFileWatcher(dir string, config phoenix.Configuration) chan phoenix.AssetEvent {
	filter := phoenix.NewEventFilterFromPatternsAndFiles(config.IgnoredFiles, config.Ignores)
	return phoenix.NewFileWatcher(dir, true, filter)
}
