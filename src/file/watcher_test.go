package file

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/radovskyb/watcher"
	"github.com/stretchr/testify/assert"

	"github.com/Shopify/themekit/src/env"
)

func TestMain(m *testing.M) {
	drainTimeout = 100 * time.Millisecond
	pollInterval = time.Nanosecond
	os.Exit(m.Run())
}

func TestNewFileWatcher(t *testing.T) {
	w := createTestWatcher(t)
	filter, _ := NewFilter(filepath.Join("_testdata", "project"), []string{"config/"}, []string{})
	assert.Equal(t, w.filter, filter)
	assert.Equal(t, w.notify, "/tmp/notifytest")
	assert.NotNil(t, w.fsWatcher)
	assert.NotNil(t, w.Events)
	watchedFiles := w.fsWatcher.WatchedFiles()
	assert.Equal(t, 14, len(watchedFiles))
}

func TestFileWatcher_Watch(t *testing.T) {
	e := &env.Env{
		Directory:    filepath.Join("_testdata", "project"),
		IgnoredFiles: []string{"config"},
	}
	w, _ := NewWatcher(e, "")
	w.Watch()

	path := filepath.Join("_testdata", "project", "assets", "application.js")
	info, _ := os.Stat(path)
	w.fsWatcher.Wait()
	w.fsWatcher.Event <- watcher.Event{Op: watcher.Create, Path: path, FileInfo: info}

	select {
	case <-w.Events:
	case <-time.After(time.Second):
		t.Error("Didnt process an event so must not be watching")
	}

	w.Stop()
}

func TestFileWatcher_WatchFsEvents(t *testing.T) {
	testcases := []struct {
		filename      string
		op            watcher.Op
		shouldReceive bool
		expectedOp    Op
	}{
		{shouldReceive: false, filename: "_testdata/project/config/settings.json", op: watcher.Write},
		{shouldReceive: true, expectedOp: Update, filename: "_testdata/project/templates/template.liquid", op: watcher.Write},
		{shouldReceive: true, expectedOp: Update, filename: "_testdata/project/templates/customers/test.liquid", op: watcher.Write},
		{shouldReceive: true, expectedOp: Remove, filename: "_testdata/project/templates/customers/test.liquid", op: watcher.Remove},
		{shouldReceive: true, expectedOp: Remove, filename: "_testdata/project/templates/customers/test.liquid", op: watcher.Rename},
	}

	w := createTestWatcher(t)
	w.Watch()
	w.Events = make(chan Event, len(testcases))
	defer w.Stop()
	for i, testcase := range testcases {
		info, _ := os.Stat(testcase.filename)
		w.fsWatcher.Event <- watcher.Event{Op: testcase.op, Path: testcase.filename, FileInfo: info}

		if testcase.shouldReceive {
			e := <-w.Events
			assert.Contains(t, testcase.filename, e.Path)
			assert.Equal(t, testcase.expectedOp, e.Op, fmt.Sprintf("got the wrong operation %v", i))
		} else {
			if !assert.False(t, len(w.Events) > 0, testcase.filename) {
				<-w.Events
			}
		}

		for len(w.Events) > 0 {
			<-w.Events
		}
	}
}

func TestFileWatcher_OnEvent(t *testing.T) {
	testcases := []struct {
		filename   string
		op         watcher.Op
		expectedOp []Op
	}{
		{expectedOp: []Op{}, filename: "_testdata/project/templates/customers", op: watcher.Write},
		{expectedOp: []Op{}, filename: "_testdata/project/config/settings.json", op: watcher.Write},
		{expectedOp: []Op{Update}, filename: "_testdata/project/templates/template.liquid", op: watcher.Write},
		{expectedOp: []Op{Update}, filename: "_testdata/project/templates/customers/test.liquid", op: watcher.Create},
		{expectedOp: []Op{Remove}, filename: "_testdata/project/templates/customers/test.liquid", op: watcher.Remove},
		{expectedOp: []Op{Remove, Update}, filename: "_testdata/project/assets/application.js.liquid -> _testdata/project/assets/application.js", op: watcher.Rename},
		{expectedOp: []Op{Remove, Update}, filename: "_testdata/project/assets/application.js.liquid -> _testdata/project/assets/application.js", op: watcher.Move},
	}

	w := createTestWatcher(t)
	w.Events = make(chan Event, len(testcases))
	defer w.Stop()
	for i, testcase := range testcases {
		_, currentPath := w.parsePath(testcase.filename)
		info, _ := os.Stat(filepath.Join("_testdata", "project", currentPath))
		w.onEvent(watcher.Event{Op: testcase.op, Path: testcase.filename, FileInfo: info})
		assert.Equal(t, len(testcase.expectedOp), len(w.Events), fmt.Sprintf("testcase: %v", i))
		for i := 0; i < len(testcase.expectedOp); i++ {
			e := <-w.Events
			assert.Contains(t, testcase.filename, e.Path)
			assert.Equal(t, testcase.expectedOp[i], e.Op)
		}
	}
}

func TestFileWatcher_debouncing(t *testing.T) {
	w := createTestWatcher(t)
	w.Events = make(chan Event, 10)
	path := filepath.Join("_testdata", "project", "templates", "customers", "test.liquid")
	info, _ := os.Stat(path)
	go func() {
		w.fsWatcher.Event <- watcher.Event{Op: watcher.Write, Path: path, FileInfo: info}
		w.fsWatcher.Event <- watcher.Event{Op: watcher.Write, Path: path, FileInfo: info}
		w.fsWatcher.Event <- watcher.Event{Op: watcher.Remove, Path: path, FileInfo: info}
	}()
	go w.watchFsEvents()
	defer w.Stop()
	time.Sleep(2 * drainTimeout)
	assert.Equal(t, 1, len(w.Events))
}

func TestFileWatcher_ParsePath(t *testing.T) {
	testcases := []struct {
		input, oldpath, currentpath string
	}{
		{"_testdata/project/assets/app.js", "", "assets/app.js"},
		{"_testdata/project/assets/app.js -> _testdata/project/assets/app.js.liquid", "assets/app.js", "assets/app.js.liquid"},
		{"not/another/path/assets/app.js", "", "not/another/path/assets/app.js"},
	}
	w := createTestWatcher(t)
	for _, testcase := range testcases {
		o, c := w.parsePath(testcase.input)
		assert.Equal(t, o, testcase.oldpath)
		assert.Equal(t, c, testcase.currentpath)
	}
}

func TestIsEventType(t *testing.T) {
	expectedOps := []watcher.Op{watcher.Write, watcher.Remove, watcher.Rename}
	refutedOps := []watcher.Op{watcher.Chmod, watcher.Create, watcher.Move}

	for _, op := range expectedOps {
		assert.True(t, isEventType(op, expectedOps...), fmt.Sprintf("%v", op))
	}

	for _, op := range refutedOps {
		assert.False(t, isEventType(op, expectedOps...), fmt.Sprintf("%v", op))
	}
}

func TestFileWatcher_StopWatching(t *testing.T) {
	w := createTestWatcher(t)
	w.Stop()
	w.Watch()
	w.Stop()
}

func TestFileWatcher_TouchNotifyFile(t *testing.T) {
	notifyPath := filepath.Join("_testdata", "notify_file")
	w, _ := NewWatcher(&env.Env{Notify: notifyPath}, "")

	os.Remove(notifyPath)
	_, err := os.Stat(notifyPath)
	assert.True(t, os.IsNotExist(err))

	w.onIdle()
	_, err = os.Stat(notifyPath)
	assert.Nil(t, err)
	// need to make the time different larger than milliseconds because windows
	// trucates the time and it will fail
	os.Chtimes(w.notify, time.Now().AddDate(0, 0, -1), time.Now().AddDate(0, 0, -1))
	info1, err := os.Stat(notifyPath)
	assert.Nil(t, err)

	w.onIdle()
	info2, err := os.Stat(notifyPath)
	assert.Nil(t, err)
	assert.NotEqual(t, info1.ModTime(), info2.ModTime())
}

func createTestWatcher(t *testing.T) *Watcher {
	e := &env.Env{
		Directory:    filepath.Join("_testdata", "project"),
		IgnoredFiles: []string{"config/"},
		Notify:       "/tmp/notifytest",
	}
	w, err := NewWatcher(e, filepath.Join("_testdata", "project", "config.yml"))
	assert.Nil(t, err)
	return w
}
