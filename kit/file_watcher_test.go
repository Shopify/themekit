package kit

import (
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/stretchr/testify/assert"

	"github.com/Shopify/themekit/kittest"
)

func TestNewFileReader(t *testing.T) {
	kittest.GenerateProject()
	defer kittest.Cleanup()
	client := ThemeClient{Config: &Configuration{Directory: kittest.FixtureProjectPath}}
	watcher, err := newFileWatcher(client, "", true, fileFilter{}, func(ThemeClient, Asset, EventType) {})
	assert.Nil(t, err)
	assert.Equal(t, true, watcher.IsWatching())
	watcher.StopWatching()
}

func TestWatchDirectory(t *testing.T) {
	kittest.GenerateProject()
	defer kittest.Cleanup()
	filter, _ := newFileFilter(kittest.FixtureProjectPath, []string{}, []string{})
	w, _ := fsnotify.NewWatcher()
	watcher := &FileWatcher{
		filter:      filter,
		mainWatcher: w,
		client:      ThemeClient{Config: &Configuration{Directory: kittest.FixtureProjectPath}},
	}
	watcher.watch()
	assert.Nil(t, watcher.mainWatcher.Remove(filepath.Join(kittest.FixtureProjectPath, "assets")))
	watcher.StopWatching()
}

func TestWatchSymlinkDirectory(t *testing.T) {
	kittest.GenerateProject()
	defer kittest.Cleanup()
	filter, _ := newFileFilter(kittest.SymlinkProjectPath, []string{}, []string{})
	w, _ := fsnotify.NewWatcher()
	watcher := &FileWatcher{
		filter:      filter,
		mainWatcher: w,
		client:      ThemeClient{Config: &Configuration{Directory: kittest.SymlinkProjectPath}},
	}
	assert.Nil(t, watcher.watch())
	assert.Nil(t, watcher.mainWatcher.Remove(filepath.Join(kittest.FixtureProjectPath, "assets")))
	watcher.StopWatching()

	os.Remove(kittest.SymlinkProjectPath)
	os.Symlink("nope", kittest.SymlinkProjectPath)
	_, err := newFileFilter(kittest.SymlinkProjectPath, []string{}, []string{})
	assert.NotNil(t, err)

	watcher.client.Config.Directory = kittest.SymlinkProjectPath
	assert.NotNil(t, watcher.watch())
	watcher.StopWatching()
}

func TestWatchConfig(t *testing.T) {
	kittest.GenerateProject()
	kittest.GenerateConfig("example.myshopify.com", true)
	defer kittest.Cleanup()
	filter, _ := newFileFilter(kittest.FixtureProjectPath, []string{}, []string{})
	w, _ := fsnotify.NewWatcher()
	watcher := &FileWatcher{
		done:          make(chan bool),
		filter:        filter,
		configWatcher: w,
	}

	err := watcher.WatchConfig("nope", make(chan bool))
	assert.NotNil(t, err)

	err = watcher.WatchConfig("config.yml", make(chan bool))
	assert.Nil(t, err)
}

func TestWatchFsEvents(t *testing.T) {
	kittest.GenerateProject()
	defer kittest.Cleanup()
	assetChan := make(chan Asset, 100)
	eventChan := make(chan fsnotify.Event)
	var wg sync.WaitGroup
	wg.Add(2)

	filter, _ := newFileFilter(kittest.FixtureProjectPath, []string{}, []string{})

	watcher := &FileWatcher{
		done:          make(chan bool),
		filter:        filter,
		mainWatcher:   &fsnotify.Watcher{Events: eventChan},
		client:        ThemeClient{Config: &Configuration{Directory: kittest.FixtureProjectPath}},
		configWatcher: &fsnotify.Watcher{Events: make(chan fsnotify.Event)},
	}

	watcher.callback = func(client ThemeClient, asset Asset, event EventType) {
		assert.Equal(t, Update, event)
		assetChan <- asset
		wg.Done()
	}

	go watcher.watchFsEvents()

	go func() {
		writes := []fsnotify.Event{
			{Name: filepath.Join(kittest.FixtureProjectPath, "templates", "template.liquid"), Op: fsnotify.Write},
			{Name: filepath.Join(kittest.FixtureProjectPath, "templates", "customers", "test.liquid"), Op: fsnotify.Write},
		}
		for _, fsEvent := range writes {
			eventChan <- fsEvent
		}
	}()

	wg.Wait()
	// test that the events are debounced
	assert.Equal(t, 2, len(assetChan))
}

func TestReloadConfig(t *testing.T) {
	kittest.GenerateProject()
	kittest.GenerateConfig("example.myshopify.com", true)
	defer kittest.Cleanup()
	reloadChan := make(chan bool, 100)

	configWatcher, _ := fsnotify.NewWatcher()
	watcher := &FileWatcher{
		done:          make(chan bool),
		mainWatcher:   &fsnotify.Watcher{Events: make(chan fsnotify.Event)},
		configWatcher: configWatcher,
	}

	watcher.callback = func(client ThemeClient, asset Asset, event EventType) {}
	err := watcher.WatchConfig("config.yml", reloadChan)
	assert.Nil(t, err)

	go watcher.watchFsEvents()
	configWatcher.Events <- fsnotify.Event{Name: "config.yml", Op: fsnotify.Write}

	_, ok := <-watcher.done
	assert.False(t, ok)
	assert.Equal(t, watcher.IsWatching(), false)
}

func TestStopWatching(t *testing.T) {
	kittest.GenerateProject()
	defer kittest.Cleanup()
	client := ThemeClient{Config: &Configuration{Directory: kittest.FixtureProjectPath}}
	watcher, err := newFileWatcher(client, "", true, fileFilter{}, func(ThemeClient, Asset, EventType) {})
	assert.Nil(t, err)
	assert.Equal(t, true, watcher.IsWatching())
	watcher.StopWatching()
	time.Sleep(50 * time.Millisecond)
	assert.Equal(t, false, watcher.IsWatching())
}

func TestOnReload(t *testing.T) {
	kittest.GenerateProject()
	kittest.GenerateConfig("example.myshopify.com", true)
	defer kittest.Cleanup()
	reloadChan := make(chan bool, 100)

	configWatcher, _ := fsnotify.NewWatcher()
	watcher := &FileWatcher{
		done:          make(chan bool),
		mainWatcher:   &fsnotify.Watcher{Events: make(chan fsnotify.Event)},
		configWatcher: configWatcher,
		client:        ThemeClient{Config: &Configuration{Directory: kittest.FixtureProjectPath}},
	}

	err := watcher.WatchConfig("config.yml", reloadChan)
	assert.Nil(t, err)
	watcher.onReload()

	assert.Equal(t, len(reloadChan), 1)
	assert.Equal(t, watcher.IsWatching(), false)
}

func TestOnEvent(t *testing.T) {
	kittest.GenerateProject()
	defer kittest.Cleanup()

	watcher := &FileWatcher{
		waitNotify:     false,
		recordedEvents: newEventMap(),
		callback:       func(client ThemeClient, asset Asset, event EventType) {},
		client:         ThemeClient{Config: &Configuration{Directory: kittest.FixtureProjectPath}},
	}

	event1 := fsnotify.Event{Name: filepath.Join(kittest.FixtureProjectPath, "templates", "template.liquid"), Op: fsnotify.Write}
	event2 := fsnotify.Event{Name: filepath.Join(kittest.FixtureProjectPath, "templates", "customers", "test.liquid"), Op: fsnotify.Write}

	assert.Equal(t, watcher.recordedEvents.Count(), 0)
	watcher.onEvent(event1)
	assert.Equal(t, watcher.recordedEvents.Count(), 1)
	watcher.onEvent(event1)
	assert.Equal(t, watcher.recordedEvents.Count(), 1)
	watcher.onEvent(event2)
	assert.Equal(t, watcher.recordedEvents.Count(), 2)
}

func TestWatchForIdle(t *testing.T) {
	notifyPath := "notifyTestFile"
	defer os.Remove(notifyPath)
	watcher := &FileWatcher{notify: notifyPath, recordedEvents: newEventMap()}
	watcher.watchForIdle()

	watcher.mutex.Lock()
	defer watcher.mutex.Unlock()
	assert.True(t, watcher.waitNotify)
}

func TestIdleMonitor(t *testing.T) {
	notifyPath := "notifyTestFile"
	defer os.Remove(notifyPath)
	watcher := &FileWatcher{notify: notifyPath, recordedEvents: newEventMap()}
	watcher.idleMonitor()

	watcher.mutex.Lock()
	defer watcher.mutex.Unlock()
	assert.False(t, watcher.waitNotify)
}

func TestTouchNotifyFile(t *testing.T) {
	kittest.GenerateProject()
	defer kittest.Cleanup()
	notifyPath := "notifyTestFile"
	defer os.Remove(notifyPath)
	watcher := &FileWatcher{notify: notifyPath}
	os.Remove(notifyPath)
	_, err := os.Stat(notifyPath)
	assert.True(t, os.IsNotExist(err))
	watcher.waitNotify = true
	watcher.touchNotifyFile()
	_, err = os.Stat(notifyPath)
	assert.False(t, os.IsNotExist(err))
	assert.False(t, watcher.waitNotify)
}

func TestHandleEvent(t *testing.T) {
	kittest.GenerateProject()
	defer kittest.Cleanup()

	writes := []struct {
		Name          string
		Event         fsnotify.Op
		ExpectedEvent EventType
	}{
		{Name: filepath.Join(kittest.FixtureProjectPath, "assets", "application.js"), Event: fsnotify.Create, ExpectedEvent: Update},
		{Name: filepath.Join(kittest.FixtureProjectPath, "assets", "application.js"), Event: fsnotify.Write, ExpectedEvent: Update},
		{Name: filepath.Join(kittest.FixtureProjectPath, "assets", "application.js"), Event: fsnotify.Remove, ExpectedEvent: Remove},
		{Name: filepath.Join(kittest.FixtureProjectPath, "assets", "application.js"), Event: fsnotify.Rename, ExpectedEvent: Remove},
	}

	for _, write := range writes {
		watcher := &FileWatcher{callback: func(client ThemeClient, asset Asset, event EventType) {
			assert.Equal(t, pathToProject(kittest.FixtureProjectPath, filepath.Join(kittest.FixtureProjectPath, "assets", "application.js")), asset.Key)
			assert.Equal(t, write.ExpectedEvent, event)
		},
			client: ThemeClient{Config: &Configuration{Directory: kittest.FixtureProjectPath}},
		}
		watcher.handleEvent(fsnotify.Event{Name: write.Name, Op: write.Event})
	}

	// make sure no error is thrown
	watcher := &FileWatcher{done: make(chan bool)}
	close(watcher.done)
	watcher.handleEvent(fsnotify.Event{})
}
