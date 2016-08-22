package commands

import (
	"fmt"
	"os"
	"time"

	"github.com/Shopify/themekit"
)

// WatchCommand watches directories for changes, and updates the remote theme
func WatchCommand(args Args, done chan bool) {
	if isSingleEnvironment(args) {
		args.ThemeClients = []themekit.ThemeClient{args.ThemeClient}
	}

	eventLog := args.EventLog

	for _, client := range args.ThemeClients {
		config := client.GetConfiguration()
		concurrency := config.Concurrency
		logEvent(message(fmt.Sprintf("Spawning %d workers for %s", concurrency, config.Domain)), eventLog)

		args.ThemeClient = client
		watchForChangesAndIssueWork(args, eventLog)
	}
}

func isSingleEnvironment(args Args) bool {
	return len(args.ThemeClients) == 0
}

func watchForChangesAndIssueWork(args Args, eventLog chan themekit.ThemeEvent) {
	client := args.ThemeClient
	config := client.GetConfiguration()
	bucket := client.LeakyBucket()
	bucket.TopUp()

	foreman := themekit.NewForeman(bucket)
	foreman.OnIdle = func() {
		if len(args.NotifyFile) > 0 {
			os.Create(args.NotifyFile)
			os.Chtimes(args.NotifyFile, time.Now(), time.Now())
		}
	}
	watcher := constructFileWatcher(args.Directory, config)
	foreman.JobQueue = watcher
	foreman.IssueWork()

	for i := 0; i < config.Concurrency; i++ {
		workerName := fmt.Sprintf("%s Worker #%d", config.Domain, i)
		go spawnWorker(workerName, foreman.WorkerQueue, client, eventLog)
	}
}

func spawnWorker(workerName string, queue chan themekit.AssetEvent, client themekit.ThemeClient, eventLog chan themekit.ThemeEvent) {
	logEvent(workerSpawnEvent(workerName), eventLog)
	for {
		asset := <-queue
		if asset.Asset().IsValid() {
			workerEvent := basicEvent{
				Title:     "FS Event",
				EventType: asset.Type().String(),
				Target:    asset.Asset().Key,
				Etype:     "fsevent",
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
	watcher, err := themekit.NewFileWatcher(dir, true, filter)
	if err != nil {
		themekit.NotifyError(err)
	}
	return watcher
}

func workerSpawnEvent(workerName string) themekit.ThemeEvent {
	return basicEvent{
		Title:     "Worker",
		Target:    workerName,
		Etype:     "basicEvent",
		EventType: "worker",
		Formatter: func(b basicEvent) string {
			return fmt.Sprintf("%s ready to upload local changes", b.Target)
		},
	}
}
