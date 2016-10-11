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
	debounceTimeout = 500 * time.Millisecond
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

type (
	fsAssetEvent struct {
		asset     theme.Asset
		eventType EventType
	}
	fileReader func(filename string) ([]byte, error)
)

// Asset ... TODO
func (f fsAssetEvent) Asset() theme.Asset {
	return f.asset
}

// Type ... TODO
func (f fsAssetEvent) Type() EventType {
	return f.eventType
}

// IsValid ... TODO
func (f fsAssetEvent) IsValid() bool {
	return f.eventType == Remove || f.asset.IsValid()
}

func (f fsAssetEvent) String() string {
	return fmt.Sprintf("%s|%s", f.asset.Key, f.eventType.String())
}

func newFileWatcher(dir string, recur bool, filter eventFilter) (chan AssetEvent, error) {
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
			Fatal(err)
		} else {
			asset = theme.Asset{}
		}
	}
	asset.Key = extractAssetKey(event.Name)
	return asset
}

func handleEvent(event fsnotify.Event) fsAssetEvent {
	var eventType EventType
	asset := fwLoadAsset(event)
	switch event.Op {
	case fsnotify.Create:
		eventType = Update
	case fsnotify.Remove:
		eventType = Remove
	}
	return fsAssetEvent{asset: asset, eventType: eventType}
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

func convertFsEvents(events chan fsnotify.Event, filter eventFilter) chan AssetEvent {
	results := make(chan AssetEvent)
	go func() {
		var currentEvent fsnotify.Event
		recordedEvents := map[string]fsnotify.Event{}
		for {
			select {
			case currentEvent = <-events:
				if currentEvent.Op&fsnotify.Chmod == fsnotify.Chmod {
					currentEvent.Op = fsnotify.Write
				}
				recordedEvents[currentEvent.Name] = currentEvent
			case <-time.After(debounceTimeout):
				for eventName, event := range recordedEvents {
					if fsevent := handleEvent(event); !filter.MatchesFilter(eventName) && fsevent.IsValid() {
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
