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
	notify   string
}

func newFileWatcher(client ThemeClient, dir, notifyFile string, recur bool, filter eventFilter, callback FileEventCallback) (*FileWatcher, error) {
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

func (watcher *FileWatcher) watchDirectory(dir string) error {
	dir = filepath.Clean(dir)
	return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() && !watcher.filter.matchesFilter(path) && path != dir {
			for _, dir := range assetLocations {
				if strings.Contains(path+"/", dir) {
					if err := watcher.watcher.Add(path); err != nil {
						return fmt.Errorf("Could not watch directory %s: %s", path, err)
					}
				}
			}
		}
		return nil
	})
}

func (watcher *FileWatcher) watchFsEvents(notifyFile string) {
	var eventLock sync.Mutex
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
			eventLock.Unlock()
		case <-time.Tick(1 * time.Second):
			eventLock.Lock()
			println(len(recordedEvents) == 0, notifyFile != "")
			if len(recordedEvents) == 0 && notifyFile != "" {
				os.Create(notifyFile)
				os.Chtimes(notifyFile, time.Now(), time.Now())
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
	filename = filepath.ToSlash(filename)

	for _, dir := range assetLocations {
		split := strings.SplitAfterN(filename, dir, 2)
		if len(split) > 1 {
			return fmt.Sprintf("%s%s", dir, split[len(split)-1])
		}
	}

	return ""
}
