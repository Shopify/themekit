package kit

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/fsnotify.v1"

	"github.com/Shopify/themekit/theme"
)

const (
	debouceTimeout = 50 * time.Millisecond
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
	WatcherFileReader fileReader = ioutil.ReadFile
)

type (
	FsAssetEvent struct {
		asset     theme.Asset
		eventType EventType
	}
	fileReader func(filename string) ([]byte, error)
)

// RestoreReader ... TODO
func RestoreReader() {
	WatcherFileReader = ioutil.ReadFile
}

// Asset ... TODO
func (f FsAssetEvent) Asset() theme.Asset {
	return f.asset
}

// Type ... TODO
func (f FsAssetEvent) Type() EventType {
	return f.eventType
}

// IsValid ... TODO
func (f FsAssetEvent) IsValid() bool {
	return f.eventType == Remove || f.asset.IsValid()
}

func (f FsAssetEvent) String() string {
	return fmt.Sprintf("%s|%s", f.asset.Key, f.eventType.String())
}

// NewFileWatcher ... TODO
func NewFileWatcher(dir string, recur bool, filter EventFilter) (chan AssetEvent, error) {
	dirsToWatch, err := findDirectoriesToWatch(dir, recur, filter.MatchesFilter)
	if err != nil {
		return nil, err
	}

	watcher, err := fsnotify.NewWatcher()
	// TODO: the watcher should be closed at the end!!
	if err != nil {
		return nil, err
	}

	for _, path := range dirsToWatch {
		if err := watcher.Add(path); err != nil {
			return nil, fmt.Errorf("Could not watch directory %s: %s", path, err)
		}
	}

	return convertFsEvents(watcher.Events, filter), nil
}

func findDirectoriesToWatch(start string, recursive bool, ignoreDirectory func(string) bool) ([]string, error) {
	var result []string
	if !recursive {
		result = append(result, start)
		return result, nil
	}

	walkFunc := func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			return nil
		}
		if ignoreDirectory(path) {
			return nil
		}
		result = append(result, path)
		return nil
	}
	if err := filepath.Walk(start, walkFunc); err != nil {
		return nil, err
	}
	return result, nil
}

func fwLoadAsset(event fsnotify.Event) theme.Asset {
	root := filepath.Dir(event.Name)
	filename := filepath.Base(event.Name)

	asset, err := theme.LoadAsset(root, filename)
	if err != nil {
		if os.IsExist(err) {
			NotifyError(err)
		} else {
			asset = theme.Asset{}
		}
	}
	asset.Key = extractAssetKey(event.Name)
	return asset
}

// HandleEvent ... TODO
func HandleEvent(event fsnotify.Event) FsAssetEvent {
	var eventType EventType
	asset := fwLoadAsset(event)
	switch event.Op {
	case fsnotify.Create:
		eventType = Update
	case fsnotify.Remove:
		eventType = Remove
	}
	return FsAssetEvent{asset: asset, eventType: eventType}
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

func convertFsEvents(events chan fsnotify.Event, filter EventFilter) chan AssetEvent {
	results := make(chan AssetEvent)
	go func() {
		var currentEvent fsnotify.Event
		recordedEvents := map[string]fsnotify.Event{}
		for {
			currentEvent = <-events
			select {
			case currentEvent = <-events:
				if currentEvent.Op == fsnotify.Chmod {
					continue
				}
				recordedEvents[currentEvent.Name] = currentEvent
			case <-time.After(debouceTimeout):
				for eventName, event := range recordedEvents {
					if fsevent := HandleEvent(event); !filter.MatchesFilter(eventName) && fsevent.IsValid() {
						results <- fsevent
					}
				}
				recordedEvents = map[string]fsnotify.Event{}
				break
			}
		}
	}()
	return results
}
