package themekit

import (
	"encoding/base64"
	"fmt"
	"gopkg.in/fsnotify.v1"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const EventTimeoutInMs int64 = 3000

var assetLocations = []string{"templates/customers/", "assets/", "config/", "layout/", "snippets/", "templates/", "locales/"}

type FsAssetEvent struct {
	asset     Asset
	eventType EventType
}

type FileReader func(filename string) ([]byte, error)

var WatcherFileReader FileReader = ioutil.ReadFile

func RestoreReader() {
	WatcherFileReader = ioutil.ReadFile
}

func (f FsAssetEvent) Asset() Asset {
	return f.asset
}

func (f FsAssetEvent) Type() EventType {
	return f.eventType
}

func (f FsAssetEvent) IsValid() bool {
	return f.eventType == Remove || f.asset.IsValid()
}

func (f FsAssetEvent) String() string {
	return fmt.Sprintf("%s|%s", f.asset.Key, f.eventType.String())
}

func NewFileWatcher(dir string, recur bool, filter EventFilter) (chan AssetEvent, error) {
	if recur {
		return watchDirRecur(dir, filter)
	} else {
		return watchDir(dir, filter)
	}
}

func fwLoadAsset(event fsnotify.Event) Asset {
	root := filepath.Dir(event.Name)
	filename := filepath.Base(event.Name)

	asset, err := LoadAsset(root, filename)
	if err != nil {
		if os.IsExist(err) {
			NotifyError(err)
		} else {
			asset = Asset{}
		}
	}
	asset.Key = extractAssetKey(event.Name)
	return asset
}

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

func ContentTypeFor(data []byte) string {
	contentType := http.DetectContentType(data)
	if strings.Contains(contentType, "text") {
		return "text"
	} else {
		return "binary"
	}
}

func Encode64(data []byte) string {
	return base64.StdEncoding.EncodeToString(data)
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
		duplicateEventTimeout := map[string]int64{}
		for {
			event := <-events

			if event.Op == fsnotify.Chmod {
				continue
			}

			if !filter.MatchesFilter(event.Name) {
				fsevent := HandleEvent(event)
				duplicateEventTimeoutKey := fsevent.String()
				timestamp := (time.Now().UnixNano() / int64(time.Millisecond))

				if duplicateEventTimeout[duplicateEventTimeoutKey] < timestamp && fsevent.IsValid() {
					duplicateEventTimeout[duplicateEventTimeoutKey] = timestamp + EventTimeoutInMs
					results <- fsevent
				}
			}
		}
	}()
	return results
}

func watchDirRecur(dir string, filter EventFilter) (results chan AssetEvent, err error) {
	results = make(chan AssetEvent)
	err = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() && !filter.MatchesFilter(path) {
			channel, _ := watchDir(path, filter)
			go func() {
				for {
					results <- <-channel
				}
			}()
		}
		return err
	})
	return
}

func watchDir(dir string, filter EventFilter) (results chan AssetEvent, err error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		results = make(chan AssetEvent)
		close(results)
		return results, err
	}
	err = watcher.Add(dir)
	if err != nil {
		results = make(chan AssetEvent)
		close(results)
	} else {
		results = convertFsEvents(watcher.Events, filter)
	}
	return
}
