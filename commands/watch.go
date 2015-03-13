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
		phoenix.NotifyError(errors.New("themeClient is not of valid type"))
	}

	if args["directory"] == nil {
		dir, _ = os.Getwd()
	} else if dir, ok = args["directory"].(string); !ok {
		phoenix.NotifyError(errors.New("directory is not of valid type"))
	}

	return Watch(client, dir)
}

func Watch(client phoenix.ThemeClient, dir string) chan bool {
	done := make(chan bool)
	eventLog := make(chan phoenix.ThemeEvent)

	config := client.GetConfiguration()

	bucket := phoenix.NewLeakyBucket(config.BucketSize, config.RefillRate, 1)
	bucket.TopUp()
	foreman := phoenix.NewForeman(bucket)
	watcher := constructFileWatcher(dir, config)
	foreman.JobQueue = watcher
	foreman.IssueWork()

	logEvent(message(fmt.Sprintf("Spawning %d workers", config.Concurrency)), eventLog)
	for i := 0; i < config.Concurrency; i++ {
		go spawnWorker(i, foreman.WorkerQueue, client, eventLog)
	}

	return done
}

func spawnWorker(workerId int, queue chan phoenix.AssetEvent, client phoenix.ThemeClient, eventLog chan phoenix.ThemeEvent) {
	logEvent(workerSpawnEvent(workerId), eventLog)
	for {
		asset := <-queue
		if asset.Asset().IsValid() {
			workerEvent := basicEvent{
				Title:     "FS Event",
				EventType: asset.Type().String(),
				Target:    asset.Asset().Key,
				etype:     "fsevent",
				Formatter: func(b basicEvent) string {
					return fmt.Sprintf("Received %s event on '%s'", b.EventType, b.Target)
				},
			}
			logEvent(workerEvent, eventLog)
			logEvent(client.Perform(asset), eventLog)
		}
	}
}

func constructFileWatcher(dir string, config phoenix.Configuration) chan phoenix.AssetEvent {
	filter := phoenix.NewEventFilterFromPatternsAndFiles(config.IgnoredFiles, config.Ignores)
	return phoenix.NewFileWatcher(dir, true, filter)
}

func workerSpawnEvent(workerId int) phoenix.ThemeEvent {
	return basicEvent{
		Title:     "Worker",
		Target:    fmt.Sprintf("%d", workerId),
		etype:     "basicEvent",
		EventType: "worker",
		Formatter: func(b basicEvent) string {
			return fmt.Sprintf("%s #%s ready to upload local changes", b.Title, b.Target)
		},
	}
}
