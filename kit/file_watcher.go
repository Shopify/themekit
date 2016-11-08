package kit

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"gopkg.in/fsnotify.v1"
)

const (
	debounceTimeout = 1 * time.Second
)

var (
	assetLocations = []string{
		"templates/customers/",
		"assets/",
		"config/",
		"layout/",
		"snippets/",
		"templates/",
		"locales/",
		"sections/",
	}
)

// FileEventCallback is the callback that is called when there is an event from
// a file watcher.
type FileEventCallback func(ThemeClient, Asset, EventType, error)

// FileWatcher is the object used to watch files for change and notify on any events,
// these events can then be passed along to kit to be sent to shopify.
type FileWatcher struct {
	done     chan bool
	client   ThemeClient
	watcher  *fsnotify.Watcher
	filter   eventFilter
	callback FileEventCallback
}

func newFileWatcher(client ThemeClient, dir string, recur bool, filter eventFilter, callback FileEventCallback) (*FileWatcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	for _, path := range findDirectoriesToWatch(dir, recur, filter.matchesFilter) {
		if err := watcher.Add(path); err != nil {
			return nil, fmt.Errorf("Could not watch directory %s: %s", path, err)
		}
	}

	newWatcher := &FileWatcher{
		done:     make(chan bool),
		client:   client,
		watcher:  watcher,
		callback: callback,
		filter:   filter,
	}

	go convertFsEvents(newWatcher)

	return newWatcher, nil
}

func convertFsEvents(watcher *FileWatcher) {
	var eventLock sync.Mutex
	recordedEvents := map[string]chan fsnotify.Event{}

	for {
		currentEvent, more := <-watcher.watcher.Events
		if !more {
			close(watcher.done)
			break
		}

		if currentEvent.Op == fsnotify.Chmod || watcher.filter.matchesFilter(currentEvent.Name) {
			continue
		}

		eventLock.Lock()
		if _, ok := recordedEvents[currentEvent.Name]; !ok {
			recordedEvents[currentEvent.Name] = make(chan fsnotify.Event)

			go func(eventChan chan fsnotify.Event, eventName string) {
				var event fsnotify.Event
				for {
					select {
					case event = <-eventChan:
					case <-time.After(debounceTimeout):
						go handleEvent(watcher, event)
						eventLock.Lock()
						delete(recordedEvents, eventName)
						eventLock.Unlock()
						return
					}
				}
			}(recordedEvents[currentEvent.Name], currentEvent.Name)
		}
		recordedEvents[currentEvent.Name] <- currentEvent
		eventLock.Unlock()
	}
}

// IsWatching will return true if the watcher is currently watching for file changes.
// it will return false if it has been stopped
func (watcher *FileWatcher) IsWatching() bool {
	select {
	case _, ok := <-watcher.done:
		return ok
	default:
		return true
	}
}

// StopWatching will stop the Filewatcher from watching it's directories and clean
// up any go routines doing work.
func (watcher *FileWatcher) StopWatching() {
	watcher.watcher.Close()
}

func handleEvent(watcher *FileWatcher, event fsnotify.Event) {
	var eventType EventType
	var err error

	switch event.Op {
	case fsnotify.Chmod, fsnotify.Create, fsnotify.Write:
		eventType = Update
	case fsnotify.Remove:
		eventType = Remove
	}

	root := filepath.Dir(event.Name)
	filename := filepath.Base(event.Name)
	asset, loadErr := loadAsset(root, filename)
	if loadErr != nil { // remove event wont load asset
		asset = Asset{}
	}

	asset.Key = extractAssetKey(event.Name)
	if asset.Key == "" {
		err = fmt.Errorf("File not in project workspace.")
		asset.Key = event.Name
	}

	watcher.callback(watcher.client, asset, eventType, err)
}

func extractAssetKey(filename string) string {
	filename = filepath.ToSlash(filename)

	for _, dir := range assetLocations {
		split := strings.SplitAfterN(filename, dir, 2)
		if len(split) > 1 {
			return fmt.Sprintf("%s%s", dir, split[len(split)-1])
		}
	}

	return ""
}

func findDirectoriesToWatch(start string, recursive bool, ignoreDirectory func(string) bool) []string {
	start = filepath.Clean(start)

	if !recursive {
		return []string{start}
	}

	result := []string{}
	filepath.Walk(start, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() && !ignoreDirectory(path) {
			result = append(result, path)
		}
		return nil
	})

	return result
}
