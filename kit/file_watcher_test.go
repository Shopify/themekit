package kit

import (
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

var (
	textFixturePath  = filepath.Join("..", "fixtures", "project", "assets", "application.js")
	watchFixturePath = filepath.Join("..", "fixtures", "project")
)

type FileWatcherTestSuite struct {
	suite.Suite
	watcher *FileWatcher
}

func (suite *FileWatcherTestSuite) TestNewFileReader() {
	watcher, err := newFileWatcher(ThemeClient{}, watchFixturePath, "", true, fileFilter{}, func(ThemeClient, Asset, EventType) {})
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), true, watcher.IsWatching())
	watcher.StopWatching()
}

func (suite *FileWatcherTestSuite) WatchDirectory() {
	eventChan := make(chan fsnotify.Event)
	filter, _ := newFileFilter(watchFixturePath, []string{}, []string{})
	newWatcher := &FileWatcher{
		done:        make(chan bool),
		filter:      filter,
		mainWatcher: &fsnotify.Watcher{Events: eventChan},
	}

	newWatcher.watchDirectory(watchFixturePath)
	path, _ := filepath.Abs(textFixturePath)
	assert.Nil(suite.T(), newWatcher.mainWatcher.Remove(path))
}

func (suite *FileWatcherTestSuite) TestWatchConfig() {
	filter, _ := newFileFilter(watchFixturePath, []string{}, []string{})
	w, _ := fsnotify.NewWatcher()
	newWatcher := &FileWatcher{
		done:          make(chan bool),
		filter:        filter,
		configWatcher: w,
	}

	err := newWatcher.WatchConfig("nope", make(chan bool))
	assert.NotNil(suite.T(), err)

	err = newWatcher.WatchConfig(goodEnvirontmentPath, make(chan bool))
	assert.Nil(suite.T(), err)
}

func (suite *FileWatcherTestSuite) TestWatchFsEvents() {
	assetChan := make(chan Asset, 100)
	eventChan := make(chan fsnotify.Event)
	var wg sync.WaitGroup
	wg.Add(2)

	filter, _ := newFileFilter(watchFixturePath, []string{}, []string{})

	newWatcher := &FileWatcher{
		done:          make(chan bool),
		filter:        filter,
		mainWatcher:   &fsnotify.Watcher{Events: eventChan},
		configWatcher: &fsnotify.Watcher{Events: make(chan fsnotify.Event)},
	}

	newWatcher.callback = func(client ThemeClient, asset Asset, event EventType) {
		assert.Equal(suite.T(), Update, event)
		assetChan <- asset
		wg.Done()
	}

	go newWatcher.watchFsEvents("")

	go func() {
		writes := []fsnotify.Event{
			{Name: filepath.Join(watchFixturePath, "templates", "template.liquid"), Op: fsnotify.Write},
			{Name: filepath.Join(watchFixturePath, "templates", "template.liquid"), Op: fsnotify.Write},
			{Name: filepath.Join(watchFixturePath, "templates", "template.liquid"), Op: fsnotify.Write},
			{Name: filepath.Join(watchFixturePath, "templates", "customers", "test.liquid"), Op: fsnotify.Write},
		}
		for _, fsEvent := range writes {
			eventChan <- fsEvent
		}
	}()

	wg.Wait()
	// test that the events are debounced
	assert.Equal(suite.T(), 2, len(assetChan))
}

func (suite *FileWatcherTestSuite) TestReloadConfig() {
	configChan := make(chan fsnotify.Event)
	reloadChan := make(chan bool, 100)

	newWatcher := &FileWatcher{
		done:          make(chan bool),
		mainWatcher:   &fsnotify.Watcher{Events: make(chan fsnotify.Event)},
		configWatcher: &fsnotify.Watcher{Events: configChan},
	}

	newWatcher.callback = func(client ThemeClient, asset Asset, event EventType) {}
	err := newWatcher.WatchConfig(goodEnvirontmentPath, reloadChan)
	assert.Nil(suite.T(), err)

	go newWatcher.watchFsEvents("")
	configChan <- fsnotify.Event{Name: goodEnvirontmentPath, Op: fsnotify.Write}

	assert.Equal(suite.T(), len(reloadChan), 1)
	assert.Equal(suite.T(), newWatcher.IsWatching(), false)
}

func (suite *FileWatcherTestSuite) TestStopWatching() {
	watcher, err := newFileWatcher(ThemeClient{}, watchFixturePath, "", true, fileFilter{}, func(ThemeClient, Asset, EventType) {})
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
	}

	var wg sync.WaitGroup
	wg.Add(len(writes))

	watcher := &FileWatcher{callback: func(client ThemeClient, asset Asset, event EventType) {
		assert.Equal(suite.T(), pathToProject(textFixturePath), asset.Key)
		wg.Done()
	}}

	for _, write := range writes {
		handleEvent(watcher, fsnotify.Event{Name: write.Name, Op: write.Event})
	}

	wg.Wait()
}

func TestFileWatcherTestSuite(t *testing.T) {
	suite.Run(t, new(FileWatcherTestSuite))
}

func clean(path string) string {
	return filepath.Join(strings.Split(filepath.Clean(path), "/")...)
}
