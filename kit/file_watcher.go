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
	debounceTimeout = 1100 * time.Millisecond
)

var (
	assetLocations = []string{
		filepath.FromSlash("templates/customers"),
		"assets",
		"config",
		"layout",
		"snippets",
		"templates",
		"locales",
		"sections",
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
	filter   fileFilter
	callback FileEventCallback
	notify   string
}

func newFileWatcher(client ThemeClient, dir, notifyFile string, recur bool, filter fileFilter, callback FileEventCallback) (*FileWatcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	newWatcher := &FileWatcher{
		done:     make(chan bool),
		client:   client,
		watcher:  watcher,
		callback: callback,
		filter:   filter,
	}

	go newWatcher.watchFsEvents(notifyFile)

	return newWatcher, newWatcher.watchDirectory(dir)
}

func (watcher *FileWatcher) watchDirectory(root string) error {
	root = filepath.Clean(root)
	return filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() && !watcher.filter.matchesFilter(path) && path != root {
			for _, dir := range assetLocations {
				if strings.HasPrefix(path, filepath.Join(root, dir, string(filepath.Separator))) {
					if err := watcher.watcher.Add(path); err != nil {
						return fmt.Errorf("Could not watch directory %s: %s", path, err)
					}
					break
				}
			}
		}
		return nil
	})
}

func (watcher *FileWatcher) watchFsEvents(notifyFile string) {
	var eventLock sync.Mutex
	notifyProcessed := false
	recordedEvents := map[string]chan fsnotify.Event{}

	for {
		select {
		case currentEvent, more := <-watcher.watcher.Events:
			if !more {
				close(watcher.done)
				return
			}

			if currentEvent.Op == fsnotify.Chmod || watcher.filter.matchesFilter(currentEvent.Name) {
				break
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
			notifyProcessed = true
			eventLock.Unlock()
		case <-time.Tick(1 * time.Second):
			eventLock.Lock()
			if notifyProcessed && len(recordedEvents) == 0 && notifyFile != "" {
				os.Create(notifyFile)
				os.Chtimes(notifyFile, time.Now(), time.Now())
				notifyProcessed = false
			}
			eventLock.Unlock()
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
	eventType := Update
	if event.Op&fsnotify.Remove == fsnotify.Remove {
		eventType = Remove
	}

	root := filepath.Dir(event.Name)
	filename := filepath.Base(event.Name)
	asset, loadErr := loadAsset(root, filename)
	if loadErr != nil { // remove event wont load asset
		asset = Asset{}
	}

	var err error
	asset.Key = extractAssetKey(event.Name)
	if asset.Key == "" {
		err = fmt.Errorf("file %s is not in project workspace", event.Name)
		asset.Key = event.Name
	}

	watcher.callback(watcher.client, asset, eventType, err)
}

func extractAssetKey(filename string) string {
	for _, dir := range assetLocations {
		split := strings.SplitAfterN(filename, dir+string(filepath.Separator), 2)
		if len(split) > 1 {
			return filepath.ToSlash(filepath.Join(dir, split[len(split)-1]))
		}
	}
	return ""
}

func assetInProject(root, filename string) bool {
	isAbs := strings.Contains(filename, root)
	filename += string(filepath.Separator)
	for _, dir := range assetLocations {
		path := filepath.Join(root, dir) + string(filepath.Separator)
		if (isAbs && strings.HasPrefix(filename, path)) || strings.HasPrefix(filename, dir+string(filepath.Separator)) {
			return true
		}
	}
	return false
}
