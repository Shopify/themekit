package kit

import (
	"os"
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
	watchFixturePath   = filepath.Join("..", "fixtures", "project")
	textFixturePath    = filepath.Join(watchFixturePath, "assets", "application.js")
	symlinkFixturePath = filepath.Join("..", "fixtures", "symlink_project")
)

type FileWatcherTestSuite struct {
	suite.Suite
	watcher *FileWatcher
}

func (suite *FileWatcherTestSuite) TestNewFileReader() {
	client := ThemeClient{Config: &Configuration{Directory: watchFixturePath}}
	watcher, err := newFileWatcher(client, "", true, fileFilter{}, func(ThemeClient, Asset, EventType) {})
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), true, watcher.IsWatching())
	watcher.StopWatching()
}

func (suite *FileWatcherTestSuite) TestWatchDirectory() {
	filter, _ := newFileFilter(watchFixturePath, []string{}, []string{})
	w, _ := fsnotify.NewWatcher()
	newWatcher := &FileWatcher{
		filter:      filter,
		mainWatcher: w,
		client:      ThemeClient{Config: &Configuration{Directory: watchFixturePath}},
	}
	newWatcher.watch()
	assert.Nil(suite.T(), newWatcher.mainWatcher.Remove(filepath.Join(watchFixturePath, "assets")))
}

func (suite *FileWatcherTestSuite) TestWatchSymlinkDirectory() {
	filter, _ := newFileFilter(symlinkFixturePath, []string{}, []string{})
	w, _ := fsnotify.NewWatcher()
	newWatcher := &FileWatcher{
		filter:      filter,
		mainWatcher: w,
		client:      ThemeClient{Config: &Configuration{Directory: symlinkFixturePath}},
	}
	newWatcher.watch()
	assert.Nil(suite.T(), newWatcher.mainWatcher.Remove(filepath.Join(watchFixturePath, "assets")))
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
		client:        ThemeClient{Config: &Configuration{Directory: watchFixturePath}},
		configWatcher: &fsnotify.Watcher{Events: make(chan fsnotify.Event)},
	}

	newWatcher.callback = func(client ThemeClient, asset Asset, event EventType) {
		assert.Equal(suite.T(), Update, event)
		assetChan <- asset
		wg.Done()
	}

	go newWatcher.watchFsEvents()

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
	reloadChan := make(chan bool, 100)

	configWatcher, _ := fsnotify.NewWatcher()
	newWatcher := &FileWatcher{
		done:          make(chan bool),
		mainWatcher:   &fsnotify.Watcher{Events: make(chan fsnotify.Event)},
		configWatcher: configWatcher,
	}

	newWatcher.callback = func(client ThemeClient, asset Asset, event EventType) {}
	err := newWatcher.WatchConfig(goodEnvirontmentPath, reloadChan)
	assert.Nil(suite.T(), err)

	go newWatcher.watchFsEvents()
	configWatcher.Events <- fsnotify.Event{Name: goodEnvirontmentPath, Op: fsnotify.Write}

	_, ok := <-newWatcher.done
	assert.False(suite.T(), ok)
	assert.Equal(suite.T(), newWatcher.IsWatching(), false)
}

func (suite *FileWatcherTestSuite) TestStopWatching() {
	client := ThemeClient{Config: &Configuration{Directory: watchFixturePath}}
	watcher, err := newFileWatcher(client, "", true, fileFilter{}, func(ThemeClient, Asset, EventType) {})
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), true, watcher.IsWatching())
	watcher.StopWatching()
	time.Sleep(50 * time.Millisecond)
	assert.Equal(suite.T(), false, watcher.IsWatching())
}

func (suite *FileWatcherTestSuite) TestOnReload() {
	reloadChan := make(chan bool, 100)

	configWatcher, _ := fsnotify.NewWatcher()
	newWatcher := &FileWatcher{
		done:          make(chan bool),
		mainWatcher:   &fsnotify.Watcher{Events: make(chan fsnotify.Event)},
		configWatcher: configWatcher,
		client:        ThemeClient{Config: &Configuration{Directory: watchFixturePath}},
	}

	err := newWatcher.WatchConfig(goodEnvirontmentPath, reloadChan)
	assert.Nil(suite.T(), err)
	newWatcher.onReload()

	assert.Equal(suite.T(), len(reloadChan), 1)
	assert.Equal(suite.T(), newWatcher.IsWatching(), false)
}

func (suite *FileWatcherTestSuite) TestOnEvent() {
	newWatcher := &FileWatcher{
		waitNotify:     false,
		recordedEvents: newEventMap(),
		callback:       func(client ThemeClient, asset Asset, event EventType) {},
		client:         ThemeClient{Config: &Configuration{Directory: watchFixturePath}},
	}

	event1 := fsnotify.Event{Name: filepath.Join(watchFixturePath, "templates", "template.liquid"), Op: fsnotify.Write}
	event2 := fsnotify.Event{Name: filepath.Join(watchFixturePath, "templates", "customers", "test.liquid"), Op: fsnotify.Write}

	assert.Equal(suite.T(), newWatcher.recordedEvents.Count(), 0)
	newWatcher.onEvent(event1)
	assert.Equal(suite.T(), newWatcher.recordedEvents.Count(), 1)
	newWatcher.onEvent(event1)
	assert.Equal(suite.T(), newWatcher.recordedEvents.Count(), 1)
	newWatcher.onEvent(event2)
	assert.Equal(suite.T(), newWatcher.recordedEvents.Count(), 2)
}

func (suite *FileWatcherTestSuite) TestTouchNotifyFile() {
	notifyPath := "notifyTestFile"
	newWatcher := &FileWatcher{
		notify: notifyPath,
	}
	assert.False(suite.T(), fileExists(notifyPath))
	newWatcher.waitNotify = true
	newWatcher.touchNotifyFile()
	assert.True(suite.T(), fileExists(notifyPath))
	assert.False(suite.T(), newWatcher.waitNotify)
	os.Remove(notifyPath)
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

func (suite *FileWatcherTestSuite) TestHandleEvent() {
	writes := []struct {
		Name          string
		Event         fsnotify.Op
		ExpectedEvent EventType
	}{
		{Name: textFixturePath, Event: fsnotify.Create, ExpectedEvent: Update},
		{Name: textFixturePath, Event: fsnotify.Write, ExpectedEvent: Update},
		{Name: textFixturePath, Event: fsnotify.Remove, ExpectedEvent: Remove},
		{Name: textFixturePath, Event: fsnotify.Rename, ExpectedEvent: Remove},
	}

	for _, write := range writes {
		watcher := &FileWatcher{callback: func(client ThemeClient, asset Asset, event EventType) {
			assert.Equal(suite.T(), pathToProject(watchFixturePath, textFixturePath), asset.Key)
			assert.Equal(suite.T(), write.ExpectedEvent, event)
		},
			client: ThemeClient{Config: &Configuration{Directory: watchFixturePath}},
		}
		watcher.handleEvent(fsnotify.Event{Name: write.Name, Op: write.Event})
	}
}

func TestFileWatcherTestSuite(t *testing.T) {
	suite.Run(t, new(FileWatcherTestSuite))
}

func clean(path string) string {
	return filepath.Join(strings.Split(filepath.Clean(path), "/")...)
}
