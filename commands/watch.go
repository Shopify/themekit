package commands

import (
	"fmt"
	"os"
	"time"

	"github.com/Shopify/themekit"
)

type WatchOptions struct {
	BasicOptions
	Directory  string
	NotifyFile string
	Clients    []themekit.ThemeClient
}

func WatchCommand(args map[string]interface{}) chan bool {
	currentDir, _ := os.Getwd()

	options := WatchOptions{Directory: currentDir}
	options.Clients = extractThemeClients(args)
	extractThemeClient(&options.Client, args)
	extractEventLog(&options.EventLog, args)
	extractString(&options.Directory, "directory", args)
	extractString(&options.NotifyFile, "notifyFile", args)

	return Watch(options)
}

func isSingleEnvironment(options WatchOptions) bool {
	return len(options.Clients) == 0
}

func Watch(options WatchOptions) chan bool {
	if isSingleEnvironment(options) {
		options.Clients = []themekit.ThemeClient{options.Client}
	}

	done := make(chan bool)
	eventLog := options.getEventLog()

	for _, client := range options.Clients {
		config := client.GetConfiguration()
		concurrency := config.Concurrency
		logEvent(message(fmt.Sprintf("Spawning %d workers for %s", concurrency, config.Domain)), eventLog)

		options.Client = client
		watchForChangesAndIssueWork(options, eventLog)
	}

	return done
}

func watchForChangesAndIssueWork(options WatchOptions, eventLog chan themekit.ThemeEvent) {
	client := options.Client
	config := client.GetConfiguration()
	bucket := client.LeakyBucket()
	bucket.TopUp()

	foreman := themekit.NewForeman(bucket)
	foreman.OnIdle = func() {
		if len(options.NotifyFile) > 0 {
			os.Create(options.NotifyFile)
			os.Chtimes(options.NotifyFile, time.Now(), time.Now())
		}
	}
	watcher := constructFileWatcher(options.Directory, config)
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
		etype:     "basicEvent",
		EventType: "worker",
		Formatter: func(b basicEvent) string {
			return fmt.Sprintf("%s ready to upload local changes", b.Target)
		},
	}
}
