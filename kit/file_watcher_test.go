package kit

import (
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gopkg.in/fsnotify.v1"
)

const (
	textFixturePath  = "../fixtures/project/whatever.txt"
	watchFixturePath = "../fixtures/project"
)

type FileWatcherTestSuite struct {
	suite.Suite
	watcher *FileWatcher
}

func (suite *FileWatcherTestSuite) TestNewFileReader() {
	watcher, err := newFileWatcher(ThemeClient{}, watchFixturePath, true, eventFilter{}, func(ThemeClient, Asset, EventType, error) {})
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), true, watcher.IsWatching())
	watcher.StopWatching()
}

func (suite *FileWatcherTestSuite) TestConvertFsEvents() {
	assetChan := make(chan Asset, 4)
	eventChan := make(chan fsnotify.Event)
	var wg sync.WaitGroup
	wg.Add(2)

	newWatcher := &FileWatcher{
		done:    make(chan bool),
		watcher: &fsnotify.Watcher{Events: eventChan},
	}

	newWatcher.callback = func(client ThemeClient, asset Asset, event EventType, err error) {
		assert.Nil(suite.T(), err)
		assert.Equal(suite.T(), Update, event)
		assetChan <- asset
		wg.Done()
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

	wg.Wait()
	// test that the events are debounced
	assert.Equal(suite.T(), 2, len(assetChan))
}

func (suite *FileWatcherTestSuite) TestCallbackEvents() {
	events := map[string]fsnotify.Event{
		watchFixturePath + "/templates/template.liquid": {Name: watchFixturePath + "/templates/template.liquid", Op: fsnotify.Write},
	}

	var wg sync.WaitGroup
	wg.Add(len(events))

	newWatcher := &FileWatcher{callback: func(client ThemeClient, asset Asset, event EventType, err error) {
		assert.Nil(suite.T(), err)
		assert.Equal(suite.T(), Asset{Key: "templates/template.liquid", Value: ""}, asset)
		assert.Equal(suite.T(), Update, event)
		wg.Done()
	}}

	callbackEvents(newWatcher, events)

	newWatcher = &FileWatcher{callback: func(client ThemeClient, asset Asset, event EventType, err error) {
		assert.NotNil(suite.T(), err)
		wg.Done()
	}}

	events = map[string]fsnotify.Event{
		"nope/template.liquid": {Name: "nope/template.liquid", Op: fsnotify.Write},
	}
	wg.Add(len(events))

	callbackEvents(newWatcher, events)

	wg.Wait()
}

func (suite *FileWatcherTestSuite) TestStopWatching() {
	watcher, err := newFileWatcher(ThemeClient{}, watchFixturePath, true, eventFilter{}, func(ThemeClient, Asset, EventType, error) {})
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), true, watcher.IsWatching())
	watcher.StopWatching()
	time.Sleep(50 * time.Millisecond)
	assert.Equal(suite.T(), false, watcher.IsWatching())
}

func (suite *FileWatcherTestSuite) TestHandleEvent() {
	writes := []fsnotify.Op{
		fsnotify.Create,
		fsnotify.Write,
		fsnotify.Remove,
	}

	var wg sync.WaitGroup
	wg.Add(len(writes))

	watcher := &FileWatcher{callback: func(client ThemeClient, asset Asset, event EventType, err error) {
		assert.Equal(suite.T(), "File not in project workspace.", err.Error())
		wg.Done()
	}}

	for _, fsEvent := range writes {
		handleEvent(watcher, fsnotify.Event{Name: textFixturePath, Op: fsEvent})
	}

	wg.Wait()
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
		clean(watchFixturePath),
		clean(watchFixturePath + "/assets"),
		clean(watchFixturePath + "/config"),
		clean(watchFixturePath + "/layout"),
		clean(watchFixturePath + "/locales"),
		clean(watchFixturePath + "/snippets"),
		clean(watchFixturePath + "/templates"),
		clean(watchFixturePath + "/templates/customers"),
	}

	files := findDirectoriesToWatch(watchFixturePath, true, func(string) bool { return false })
	assert.Equal(suite.T(), expected, files)

	files = findDirectoriesToWatch(watchFixturePath, false, func(string) bool { return false })
	assert.Equal(suite.T(), []string{watchFixturePath}, files)
}

func TestFileWatcherTestSuite(t *testing.T) {
	suite.Run(t, new(FileWatcherTestSuite))
}

func clean(path string) string {
	return filepath.Clean(path)
}
