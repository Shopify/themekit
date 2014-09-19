package main

import (
	"fmt"
	"github.com/csaunders/phoenix"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gopkg.in/fsnotify.v1"
	"testing"
)

type FileWatcherSuite struct {
	suite.Suite
}

func (s *FileWatcherSuite) TearDownTestSuite() {
	RestoreReader()
}

func (s *FileWatcherSuite) TestThatLoadAssetProperlyExtractsTheAssetKey() {
	var event fsnotify.Event
	WatcherFileReader = func(filename string) ([]byte, error) {
		assert.Equal(s.T(), event.Name, filename)
		return []byte("hello"), nil
	}

	for _, key := range assetLocations {
		event = fsnotify.Event{Name: fmt.Sprintf("/home/gopher/themes/simple/%s/thing.css", key)}
		asset := LoadAsset(event)
		assert.Equal(s.T(), asset.Key, fmt.Sprintf("%s/thing.css", key))
		assert.Equal(s.T(), "hello", asset.Value)
	}
}

func (s *FileWatcherSuite) TestHandleEventConversFSNotifyEventsIntoAssetEvents() {
	WatcherFileReader = func(filename string) ([]byte, error) {
		return []byte("hello"), nil
	}
	writes := map[fsnotify.Op]phoenix.EventType{
		fsnotify.Chmod:  phoenix.Update,
		fsnotify.Create: phoenix.Update,
		fsnotify.Remove: phoenix.Remove,
	}
	for fsEvent, phoenixEvent := range writes {
		event := fsnotify.Event{Name: "assets/whatever.txt", Op: fsEvent}
		assetEvent := HandleEvent(event)
		assert.Equal(s.T(), phoenixEvent, assetEvent.Type())
	}
}

func TestFileWatcherSuite(t *testing.T) {
	suite.Run(t, new(FileWatcherSuite))
}
