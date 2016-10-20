package kit

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gopkg.in/fsnotify.v1"

	"github.com/Shopify/themekit/theme"
)

const (
	textFixturePath  = "../fixtures/file_watcher/whatever.txt"
	watchFixturePath = "../fixtures/file_watcher"
)

type FileWatcherTestSuite struct {
	suite.Suite
	watcher *FileWatcher
}

func (suite *FileWatcherTestSuite) TestNewFileReader() {
	watcher, err := newFileWatcher(ThemeClient{}, watchFixturePath, true, eventFilter{}, func(ThemeClient, AssetEvent, error) {})
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), true, watcher.IsWatching())
	watcher.StopWatching()
}

func (suite *FileWatcherTestSuite) TestConvertFsEvents() {
	results := []AssetEvent{}
	callback := func(client ThemeClient, event AssetEvent, err error) {
		results = append(results, event)
		assert.Nil(suite.T(), err)
	}
	eventChan := make(chan fsnotify.Event)

	newWatcher := &FileWatcher{
		done:     make(chan bool),
		watcher:  &fsnotify.Watcher{Events: eventChan},
		callback: callback,
	}

	go convertFsEvents(newWatcher)

	go func() {
		writes := []fsnotify.Event{
			{Name: watchFixturePath + "/templates/template.liquid", Op: fsnotify.Write},
			{Name: watchFixturePath + "/templates/template.liquid", Op: fsnotify.Write},
			{Name: watchFixturePath + "/templates/template.liquid", Op: fsnotify.Write},
			{Name: watchFixturePath + "/templates/customers/test.liquid", Op: fsnotify.Write},
		}
		for _, fsEvent := range writes {
			eventChan <- fsEvent
		}
		close(eventChan)
	}()

	func() {
		for {
			select {
			case _, ok := <-newWatcher.done:
				if !ok {
					return
				}
			}
		}
	}()

	// test that the events are debounced
	assert.Equal(suite.T(), 2, len(results))
}

func (suite *FileWatcherTestSuite) TestCallbackEvents() {
	newWatcher := &FileWatcher{callback: func(client ThemeClient, event AssetEvent, err error) {
		assert.Nil(suite.T(), err)
		assert.Equal(suite.T(), AssetEvent{Asset: theme.Asset{Key: "templates/template.liquid", Value: ""}, Type: Update}, event)
	}}
	callbackEvents(newWatcher, map[string]fsnotify.Event{
		watchFixturePath + "/templates/template.liquid": {Name: watchFixturePath + "/templates/template.liquid", Op: fsnotify.Write},
	})

	newWatcher = &FileWatcher{callback: func(client ThemeClient, event AssetEvent, err error) {
		assert.NotNil(suite.T(), err)
	}}
	callbackEvents(newWatcher, map[string]fsnotify.Event{
		"nope/template.liquid": {Name: "nope/template.liquid", Op: fsnotify.Write},
	})
}

func (suite *FileWatcherTestSuite) TestStopWatching() {
	watcher, err := newFileWatcher(ThemeClient{}, watchFixturePath, true, eventFilter{}, func(ThemeClient, AssetEvent, error) {})
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), true, watcher.IsWatching())
	watcher.StopWatching()
	time.Sleep(50 * time.Millisecond)
	assert.Equal(suite.T(), false, watcher.IsWatching())
}

func (suite *FileWatcherTestSuite) TestHandleEvent() {
	_, err := handleEvent(fsnotify.Event{Name: "gone/oath", Op: fsnotify.Remove})
	assert.NotNil(suite.T(), err)

	writes := map[fsnotify.Op]EventType{
		fsnotify.Chmod:  Update,
		fsnotify.Create: Update,
		fsnotify.Write:  Update,
		fsnotify.Remove: Remove,
	}
	for fsEvent, themekitEvent := range writes {
		event := fsnotify.Event{Name: textFixturePath, Op: fsEvent}
		assetEvent, err := handleEvent(event)
		assert.Equal(suite.T(), themekitEvent, assetEvent.Type)
		assert.Equal(suite.T(), "File not in project workspace.", err.Error())
	}
}

func (suite *FileWatcherTestSuite) TestExtractAssetKey() {
	tests := map[string]string{
		textFixturePath:                                 "",
		"/long/path/to/config.yml":                      "",
		"/long/path/to/assets/logo.png":                 "assets/logo.png",
		"/long/path/to/templates/customers/test.liquid": "templates/customers/test.liquid",
		"/long/path/to/config/test.liquid":              "config/test.liquid",
		"/long/path/to/layout/test.liquid":              "layout/test.liquid",
		"/long/path/to/snippets/test.liquid":            "snippets/test.liquid",
		"/long/path/to/templates/test.liquid":           "templates/test.liquid",
		"/long/path/to/locales/test.liquid":             "locales/test.liquid",
		"/long/path/to/sections/test.liquid":            "sections/test.liquid",
	}
	for input, expected := range tests {
		assert.Equal(suite.T(), expected, extractAssetKey(input))
	}
}

func (suite *FileWatcherTestSuite) TestfindDirectoriesToWatch() {
	expected := []string{
		watchFixturePath,
		watchFixturePath + "/assets",
		watchFixturePath + "/config",
		watchFixturePath + "/layout",
		watchFixturePath + "/locales",
		watchFixturePath + "/snippets",
		watchFixturePath + "/templates",
		watchFixturePath + "/templates/customers",
	}

	files := findDirectoriesToWatch(watchFixturePath, true, func(string) bool { return false })
	assert.Equal(suite.T(), expected, files)

	files = findDirectoriesToWatch(watchFixturePath, false, func(string) bool { return false })
	assert.Equal(suite.T(), []string{watchFixturePath}, files)
}

func TestFileWatcherTestSuite(t *testing.T) {
	suite.Run(t, new(FileWatcherTestSuite))
}
