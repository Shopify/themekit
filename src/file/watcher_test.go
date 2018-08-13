package file

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/stretchr/testify/assert"

	"github.com/Shopify/themekit/src/env"
)

func TestNewFileWatcher(t *testing.T) {
	e := env.Env{
		Directory:    "/tmp",
		IgnoredFiles: []string{"*.png"},
		Notify:       "/tmp/notify",
	}
	filter, _ := NewFilter(e.Directory, e.IgnoredFiles, []string{})
	watcher, err := NewWatcher(&e, "")
	assert.Nil(t, err)
	assert.Equal(t, watcher.filter, filter)
	assert.Equal(t, watcher.notify, e.Notify)
}

func TestFileWatcher_WatchDirectory(t *testing.T) {
	e := env.Env{
		Directory:    filepath.Join("_testdata", "project"),
		IgnoredFiles: []string{"config"},
	}

	watcher, _ := NewWatcher(&e, "")
	watcher.Watch()
	assert.NotNil(t, watcher.fsWatcher)
	assert.NotNil(t, watcher.events)
	assert.NotNil(t, watcher.fsWatcher.Remove(filepath.Join("_testdata", "project", "config")))
	assert.Nil(t, watcher.fsWatcher.Remove(filepath.Join("_testdata", "project", "assets")))
	watcher.Stop()
}

func TestFileWatcher_WatchFsEvents(t *testing.T) {
	watcher, _ := NewWatcher(&env.Env{Directory: "_testdata/project", IgnoredFiles: []string{"config/settings.json"}}, "")
	watcher.events = make(chan Event)
	events := make(chan fsnotify.Event)
	go watcher.watchFsEvents(events, func(t time.Duration, events, complete chan fsnotify.Event) { complete <- (<-events) })

	testcases := []struct {
		event         fsnotify.Event
		shouldReceive bool
		expectedOp    Op
	}{
		{shouldReceive: false, event: fsnotify.Event{Name: "_testdata/project/config/settings.json", Op: fsnotify.Write}},
		{shouldReceive: false, event: fsnotify.Event{Name: "_testdata/project/templates/foo.liquid", Op: fsnotify.Chmod}},
		{shouldReceive: true, expectedOp: Update, event: fsnotify.Event{Name: "_testdata/project/templates/template.liquid", Op: fsnotify.Write}},
		{shouldReceive: true, expectedOp: Update, event: fsnotify.Event{Name: "_testdata/project/templates/customers/test.liquid", Op: fsnotify.Write}},
		{shouldReceive: true, expectedOp: Remove, event: fsnotify.Event{Name: "_testdata/project/templates/customers/test.liquid", Op: fsnotify.Remove}},
		{shouldReceive: true, expectedOp: Remove, event: fsnotify.Event{Name: "_testdata/project/templates/customers/test.liquid", Op: fsnotify.Rename}},
	}

	for _, testcase := range testcases {
		events <- testcase.event
		if testcase.shouldReceive {
			e := <-watcher.events
			assert.Contains(t, testcase.event.Name, e.Path)
			assert.Equal(t, testcase.expectedOp, e.Op)
		} else {
			assert.False(t, len(watcher.events) > 0, testcase.event.Name)
		}
	}

	// Shutdown sequence works
	close(events)
	_, ok := <-watcher.events
	assert.False(t, ok)
}

func TestFileWatcher_StopWatching(t *testing.T) {
	watcher, err := NewWatcher(&env.Env{Directory: "_testdata/project"}, "")
	assert.Nil(t, err)
	watcher.Stop()
	watcher.Watch()
	watcher.Stop()
	_, ok := <-watcher.events
	assert.False(t, ok)
}

func TestFileWatcher_TouchNotifyFile(t *testing.T) {
	notifyPath := filepath.Join("_testdata", "notify_file")
	watcher, _ := NewWatcher(&env.Env{Notify: notifyPath}, "")

	os.Remove(notifyPath)
	_, err := os.Stat(notifyPath)
	assert.True(t, os.IsNotExist(err))

	watcher.onIdle()
	_, err = os.Stat(notifyPath)
	assert.Nil(t, err)
	// need to make the time different larger than milliseconds because windows
	// trucates the time and it will fail
	os.Chtimes(watcher.notify, time.Now().AddDate(0, 0, -1), time.Now().AddDate(0, 0, -1))
	info1, err := os.Stat(notifyPath)
	assert.Nil(t, err)

	watcher.onIdle()
	info2, err := os.Stat(notifyPath)
	assert.Nil(t, err)
	assert.NotEqual(t, info1.ModTime(), info2.ModTime())
}

func TestDebounce(t *testing.T) {
	writes := []fsnotify.Event{
		{Name: "_testdata/project/assets/application.js", Op: fsnotify.Create},
		{Name: "_testdata/project/assets/application.js", Op: fsnotify.Write},
		{Name: "_testdata/project/assets/application.js", Op: fsnotify.Remove},
		{Name: "_testdata/project/assets/application.js", Op: fsnotify.Rename},
	}

	events := make(chan fsnotify.Event)
	complete := make(chan fsnotify.Event)
	go debounce(time.Millisecond, events, complete)

	for _, write := range writes {
		events <- write
	}

	e := <-complete
	assert.Equal(t, e.Name, "_testdata/project/assets/application.js")
	assert.Equal(t, e.Op, fsnotify.Rename)
}
