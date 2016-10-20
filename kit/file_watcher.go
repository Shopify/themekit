package kit

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/fsnotify.v1"

	"github.com/Shopify/themekit/theme"
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
type FileEventCallback func(ThemeClient, theme.Asset, EventType, error)

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
	var currentEvent fsnotify.Event
	var more bool
	recordedEvents := map[string]fsnotify.Event{}
	for {
		select {
		case currentEvent, more = <-watcher.watcher.Events:
			if !more {
				callbackEvents(watcher, recordedEvents)
				close(watcher.done)
				return
			}
			if currentEvent.Op != fsnotify.Chmod {
				recordedEvents[currentEvent.Name] = currentEvent
			}
		case <-time.After(debounceTimeout):
			callbackEvents(watcher, recordedEvents)
			recordedEvents = map[string]fsnotify.Event{}
		}
	}
}

func callbackEvents(watcher *FileWatcher, recordedEvents map[string]fsnotify.Event) {
	for eventName, event := range recordedEvents {
		if !watcher.filter.matchesFilter(eventName) {
			go handleEvent(watcher, event)
		}
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

	switch event.Op {
	case fsnotify.Chmod, fsnotify.Create, fsnotify.Write:
		eventType = Update
	case fsnotify.Remove:
		eventType = Remove
	}

	root := filepath.Dir(event.Name)
	filename := filepath.Base(event.Name)
	asset, err := theme.LoadAsset(root, filename)
	if err == nil {
		asset.Key = extractAssetKey(event.Name)
		if asset.Key == "" {
			err = fmt.Errorf("File not in project workspace.")
		}
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
