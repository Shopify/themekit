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

type fileReader func(filename string) ([]byte, error)

func newFileWatcher(client ThemeClient, dir string, recur bool, filter eventFilter, callback func(ThemeClient, AssetEvent, error)) error {
	dirsToWatch, err := findDirectoriesToWatch(dir, recur, filter.matchesFilter)
	if err != nil {
		return err
	}

	watcher, err := fsnotify.NewWatcher()
	// TODO: the watcher should be closed at the end!!
	if err != nil {
		return err
	}

	for _, path := range dirsToWatch {
		if err := watcher.Add(path); err != nil {
			return fmt.Errorf("Could not watch directory %s: %s", path, err)
		}
	}

	convertFsEvents(client, watcher.Events, filter, callback)

	return nil
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

func handleEvent(event fsnotify.Event) (AssetEvent, error) {
	var eventType EventType
	root := filepath.Dir(event.Name)
	filename := filepath.Base(event.Name)
	asset, err := theme.LoadAsset(root, filename)
	if err != nil {
		return AssetEvent{}, err
	}
	asset.Key = extractAssetKey(event.Name)

	switch event.Op {
	case fsnotify.Create:
		eventType = Update
	case fsnotify.Write:
		eventType = Update
	case fsnotify.Remove:
		eventType = Remove
	}

	return AssetEvent{
		Asset: asset,
		Type:  eventType,
	}, nil
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

func convertFsEvents(client ThemeClient, events chan fsnotify.Event, filter eventFilter, callback func(ThemeClient, AssetEvent, error)) {
	go func(client ThemeClient) {
		var currentEvent fsnotify.Event
		recordedEvents := map[string]fsnotify.Event{}
		for {
			select {
			case currentEvent = <-events:
				if currentEvent.Op != fsnotify.Chmod {
					recordedEvents[currentEvent.Name] = currentEvent
				}
			case <-time.After(debounceTimeout):
				for eventName, event := range recordedEvents {
					if !filter.matchesFilter(eventName) {
						event, err := handleEvent(event)
						callback(client, event, err)
					}
				}
				recordedEvents = map[string]fsnotify.Event{}
				break
			}
		}
	}(client)
}
