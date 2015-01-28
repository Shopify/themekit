package main

import (
	"encoding/base64"
	"fmt"
	"github.com/csaunders/phoenix"
	"gopkg.in/fsnotify.v1"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const EventTimeoutInMs int64 = 3000

var assetLocations = []string{"templates/customers", "assets", "config", "layout", "snippets", "templates"}

type FsAssetEvent struct {
	asset     phoenix.Asset
	eventType phoenix.EventType
}

type FileReader func(filename string) ([]byte, error)

var WatcherFileReader FileReader = ioutil.ReadFile

func RestoreReader() {
	WatcherFileReader = ioutil.ReadFile
}

func (f FsAssetEvent) Asset() phoenix.Asset {
	return f.asset
}

func (f FsAssetEvent) Type() phoenix.EventType {
	return f.eventType
}

func (f FsAssetEvent) IsValid() bool {
	return f.eventType == phoenix.Remove || f.asset.IsValid()
}

func (f FsAssetEvent) String() string {
	return fmt.Sprintf("%s|%s", f.asset.Key, f.eventType.String())
}

func NewFileWatcher(dir string, recur bool, filter phoenix.EventFilter) (processor chan phoenix.AssetEvent) {
	if recur {
		processor, _ = watchDirRecur(dir, filter)
	} else {
		processor, _ = watchDir(dir, filter)
	}
	return
}

func LoadAsset(event fsnotify.Event) phoenix.Asset {
	root := filepath.Dir(event.Name)
	fileParentDir := filepath.Base(root)
	filename := filepath.Base(event.Name)

	asset, err := phoenix.LoadAsset(root, filename)
	if err != nil {
		if os.IsExist(err) {
			phoenix.HaltAndCatchFire(err)
		} else {
			asset = phoenix.Asset{}
		}
	}
	asset.Key = fmt.Sprintf("%s/%s", fileParentDir, filename)
	return asset
}

func HandleEvent(event fsnotify.Event) FsAssetEvent {
	var eventType phoenix.EventType
	asset := LoadAsset(event)
	switch event.Op {
	case fsnotify.Create, fsnotify.Chmod:
		eventType = phoenix.Update
	case fsnotify.Remove:
		eventType = phoenix.Remove
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
	for _, dir := range assetLocations {
		split := strings.SplitAfterN(filename, dir, 2)
		if len(split) > 1 {
			return fmt.Sprintf("%s%s", dir, split[len(split)-1])
		}
	}
	return ""
}

func convertFsEvents(events chan fsnotify.Event, filter phoenix.EventFilter) chan phoenix.AssetEvent {
	results := make(chan phoenix.AssetEvent)
	go func() {
		duplicateEventTimeout := map[string]int64{}
		for {
			event := <-events

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

func watchDirRecur(dir string, filter phoenix.EventFilter) (results chan phoenix.AssetEvent, err error) {
	results = make(chan phoenix.AssetEvent)
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

func watchDir(dir string, filter phoenix.EventFilter) (results chan phoenix.AssetEvent, err error) {
	watcher, err := fsnotify.NewWatcher()
	err = watcher.Add(dir)
	if err != nil {
		results = make(chan phoenix.AssetEvent)
		close(results)
	} else {
		results = convertFsEvents(watcher.Events, filter)
	}
	return
}
