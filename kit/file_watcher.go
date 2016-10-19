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
	debounceTimeout = 1000 * time.Millisecond
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

// FileWatcher is the object used to watch files for change and notify on any events,
// these events can then be passed along to kit to be sent to shopify.
type FileWatcher struct {
	watching bool
	client   ThemeClient
	watcher  *fsnotify.Watcher
	filter   eventFilter
	callback func(ThemeClient, AssetEvent, error)
}

func newFileWatcher(client ThemeClient, dir string, recur bool, filter eventFilter, callback func(ThemeClient, AssetEvent, error)) (*FileWatcher, error) {
	dirsToWatch, err := findDirectoriesToWatch(dir, recur, filter.matchesFilter)
	if err != nil {
		return nil, err
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	for _, path := range dirsToWatch {
		if err := watcher.Add(path); err != nil {
			return nil, fmt.Errorf("Could not watch directory %s: %s", path, err)
		}
	}

	newWatcher := &FileWatcher{
		client:   client,
		watcher:  watcher,
		callback: callback,
		filter:   filter,
	}

	go newWatcher.convertFsEvents()

	return newWatcher, nil
}

func (watcher *FileWatcher) convertFsEvents() {
	var currentEvent fsnotify.Event
	var more bool
	watcher.watching = true
	recordedEvents := map[string]fsnotify.Event{}
	for {
		select {
		case currentEvent, more = <-watcher.watcher.Events:
			if !more {
				watcher.watching = false
				return
			}
			if currentEvent.Op != fsnotify.Chmod {
				recordedEvents[currentEvent.Name] = currentEvent
			}
		case <-time.After(debounceTimeout):
			for eventName, event := range recordedEvents {
				if !watcher.filter.matchesFilter(eventName) {
					event, err := handleEvent(event)
					watcher.callback(watcher.client, event, err)
				}
			}
			recordedEvents = map[string]fsnotify.Event{}
			break
		}
	}
}

// IsWatching will return true if the watcher is currently watching for file changes.
// it will return false if it has been stopped
func (watcher *FileWatcher) IsWatching() {
	watcher.watcher.Close()
}

// StopWatching will stop the Filewatcher from watching it's directories and clean
// up any go routines doing work.
func (watcher *FileWatcher) StopWatching() {
	watcher.watcher.Close()
}

func handleEvent(event fsnotify.Event) (AssetEvent, error) {
	var eventType EventType
	root := filepath.Dir(event.Name)
	filename := filepath.Base(event.Name)
	asset, err := theme.LoadAsset(root, filename)
	if err != nil {
		return AssetEvent{}, err
	}

	asset.Key = extractAssetKey(event.Name)
	if asset.Key == "" {
		err = fmt.Errorf("File not in project workspace.")
	}

	switch event.Op {
	case fsnotify.Chmod, fsnotify.Create, fsnotify.Write:
		eventType = Update
	case fsnotify.Remove:
		eventType = Remove
	}

	return AssetEvent{
		Asset: asset,
		Type:  eventType,
	}, err
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

func findDirectoriesToWatch(start string, recursive bool, ignoreDirectory func(string) bool) ([]string, error) {
	if !recursive {
		return []string{start}, nil
	}

	result := []string{}
	err := filepath.Walk(start, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() || ignoreDirectory(path) {
			return nil
		}
		result = append(result, path)
		return nil
	})

	return result, err
}
