package commands

import (
	"fmt"
	"github.com/csaunders/themekit"
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

	bucket := themekit.NewLeakyBucket(config.BucketSize, config.RefillRate, 1)
	bucket.TopUp()
	foreman := themekit.NewForeman(bucket)
	watcher := constructFileWatcher(options.Directory, config)
	foreman.JobQueue = watcher
	foreman.IssueWork()

	logEvent(message(fmt.Sprintf("Spawning %d workers", config.Concurrency)), eventLog)
	for i := 0; i < config.Concurrency; i++ {
		go spawnWorker(i, foreman.WorkerQueue, client, eventLog)
	}

	return done
}

func spawnWorker(workerId int, queue chan themekit.AssetEvent, client themekit.ThemeClient, eventLog chan themekit.ThemeEvent) {
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
						themekit.GreenText(b.EventType),
						themekit.BlueText(b.Target),
					)
				},
			}
			logEvent(workerEvent, eventLog)
			logEvent(client.Perform(asset), eventLog)
		}
	}
}

func constructFileWatcher(dir string, config themekit.Configuration) chan themekit.AssetEvent {
	filter := themekit.NewEventFilterFromPatternsAndFiles(config.IgnoredFiles, config.Ignores)
	return themekit.NewFileWatcher(dir, true, filter)
}

func workerSpawnEvent(workerId int) themekit.ThemeEvent {
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
