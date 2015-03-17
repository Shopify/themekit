package commands

import (
	"fmt"
	"github.com/csaunders/phoenix"
	"os"
)

type WatchOptions struct {
	BasicOptions
	Directory string
}

func WatchCommand(args map[string]interface{}) chan bool {
	currentDir, _ := os.Getwd()

	options := WatchOptions{Directory: currentDir}
	extractThemeClient(&options.Client, args)
	extractEventLog(&options.EventLog, args)
	extractString(&options.Directory, "directory", args)

	return Watch(options)
}

func Watch(options WatchOptions) chan bool {
	done := make(chan bool)
	eventLog := options.getEventLog()
	client := options.Client

	config := client.GetConfiguration()

	bucket := phoenix.NewLeakyBucket(config.BucketSize, config.RefillRate, 1)
	bucket.TopUp()
	foreman := phoenix.NewForeman(bucket)
	watcher := constructFileWatcher(options.Directory, config)
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
					return fmt.Sprintf(
						"Received %s event on %s",
						phoenix.GreenText(b.EventType),
						phoenix.BlueText(b.Target),
					)
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
