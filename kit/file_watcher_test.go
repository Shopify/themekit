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
	textFixturePath  = "../fixtures/project/assets/application.js"
	watchFixturePath = "../fixtures/project"
)

type FileWatcherTestSuite struct {
	suite.Suite
	watcher *FileWatcher
}

func (suite *FileWatcherTestSuite) TestNewFileReader() {
	watcher, err := newFileWatcher(ThemeClient{}, watchFixturePath, "", true, fileFilter{}, func(ThemeClient, Asset, EventType, error) {})
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), true, watcher.IsWatching())
	watcher.StopWatching()
}

func (suite *FileWatcherTestSuite) TestWatchFsEvents() {
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

	go newWatcher.watchFsEvents("")

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

func (suite *FileWatcherTestSuite) TestStopWatching() {
	watcher, err := newFileWatcher(ThemeClient{}, watchFixturePath, "", true, fileFilter{}, func(ThemeClient, Asset, EventType, error) {})
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), true, watcher.IsWatching())
	watcher.StopWatching()
	time.Sleep(50 * time.Millisecond)
	assert.Equal(suite.T(), false, watcher.IsWatching())
}

func (suite *FileWatcherTestSuite) TestHandleEvent() {
	writes := []struct {
		Name  string
		Event fsnotify.Op
	}{
		{Name: textFixturePath, Event: fsnotify.Create},
		{Name: textFixturePath, Event: fsnotify.Write},
		{Name: textFixturePath, Event: fsnotify.Remove},
		{Name: "../fixtures/project/whatever.txt", Event: fsnotify.Write},
	}

	var wg sync.WaitGroup
	wg.Add(len(writes))

	watcher := &FileWatcher{callback: func(client ThemeClient, asset Asset, event EventType, err error) {
		if err != nil {
			assert.Equal(suite.T(), "file ../fixtures/project/whatever.txt is not in project workspace", err.Error())
			assert.Equal(suite.T(), "../fixtures/project/whatever.txt", asset.Key)
		} else {
			assert.Equal(suite.T(), extractAssetKey(textFixturePath), asset.Key)
		}
		wg.Done()
	}}

	for _, write := range writes {
		handleEvent(watcher, fsnotify.Event{Name: write.Name, Op: write.Event})
	}

	wg.Wait()
}

func (suite *FileWatcherTestSuite) TestExtractAssetKey() {
	tests := map[string]string{
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

func TestFileWatcherTestSuite(t *testing.T) {
	suite.Run(t, new(FileWatcherTestSuite))
}

func clean(path string) string {
	return filepath.Clean(path)
}
