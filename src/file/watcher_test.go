package file

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
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
	assert.NotNil(t, w.fsWatcher)
	assert.NotNil(t, w.Events)
}

func TestFileWatcher_Watch(t *testing.T) {
	e := &env.Env{
		Directory:    filepath.Join("_testdata", "project"),
		IgnoredFiles: []string{"config"},
	}
	w, _ := NewWatcher(e, "", map[string]string{})
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

func TestFileWatcher_NoEventIfFileDidntChange(t *testing.T) {
	e := &env.Env{
		Directory:    filepath.Join("_testdata", "project"),
		IgnoredFiles: []string{"config"},
	}
	path := filepath.Join("_testdata", "project", "assets", "application.js")
	shortPath := filepath.Join("assets", "application.js")
	currentChecksum, _ := fileChecksum(e.Directory, shortPath)

	w, _ := NewWatcher(e, "", map[string]string{
		shortPath: currentChecksum,
	})
	w.Watch()

	info, _ := os.Stat(path)
	w.fsWatcher.Wait()
	w.fsWatcher.Event <- watcher.Event{Op: watcher.Write, Path: path, FileInfo: info}

	evt := <-w.Events
	assert.Equal(t, shortPath, evt.Path)
	assert.Equal(t, Skip, evt.Op)
	w.Stop()
}

func TestFileWatcher_filterHook(t *testing.T) {
	testcases := []struct {
		skip     bool
		filename string
	}{
		{filename: "_testdata/project/config/settings.json", skip: true},
		{filename: "_testdata/project/config.yml", skip: false},
		{filename: "_testdata/project/templates/template.liquid", skip: false},
		{filename: "_testdata/project/templates", skip: false},
		{filename: "_testdata", skip: true},
		{filename: "_testdata/project/javascripts/app.js", skip: true},
	}

	hook := createTestFilterHook(t)
	for i, testcase := range testcases {
		info, _ := os.Stat(testcase.filename)
		result := hook(info, testcase.filename)
		if testcase.skip {
			assert.Equal(t, result, watcher.ErrSkip, fmt.Sprintf("testcase: %v", i))
		} else {
			assert.Nil(t, result, fmt.Sprintf("testcase: %v", i))
		}
	}
}

func TestFileWatcher_WatchFsEvents(t *testing.T) {
	testcases := []struct {
		filename   string
		op         watcher.Op
		expectedOp Op
	}{
		{filename: "_testdata/project/config/settings.json", op: watcher.Write},
		{expectedOp: Update, filename: "_testdata/project/templates/template.liquid", op: watcher.Write},
		{expectedOp: Update, filename: "_testdata/project/templates/customers/test.liquid", op: watcher.Write},
		{expectedOp: Remove, filename: "_testdata/project/templates/customers/test.liquid", op: watcher.Remove},
	}

	for i, testcase := range testcases {
		w := createTestWatcher(t)
		w.Watch()
		w.Events = make(chan Event, len(testcases))
		info, _ := os.Stat(testcase.filename)
		w.fsWatcher.Event <- watcher.Event{Op: testcase.op, Path: testcase.filename, FileInfo: info}

		e := <-w.Events
		assert.Contains(t, testcase.filename, e.Path)
		assert.Equal(t, testcase.expectedOp, e.Op, fmt.Sprintf("got the wrong operation %v", i))
		w.Stop()
	}
}

type OptSlice []Op

func (p OptSlice) Len() int           { return len(p) }
func (p OptSlice) Less(i, j int) bool { return p[i] < p[j] }
func (p OptSlice) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

func TestFileWatcher_OnEvent(t *testing.T) {
	testcases := []struct {
		filename, oldname string
		op                watcher.Op
		expectedOp        OptSlice
	}{
		{expectedOp: OptSlice{Update}, filename: "_testdata/project/templates/template.liquid", op: watcher.Write},
		{expectedOp: OptSlice{Update}, filename: "_testdata/project/templates/customers/test.liquid", op: watcher.Create},
		{expectedOp: OptSlice{}, filename: "_testdata/project/config", op: watcher.Create},
		{expectedOp: OptSlice{Remove}, filename: "_testdata/project/templates/customers/test.liquid", op: watcher.Remove},
		{expectedOp: OptSlice{Remove, Update}, filename: "_testdata/project/assets/application.js.liquid", oldname: "_testdata/project/assets/application.js", op: watcher.Rename},
		{expectedOp: OptSlice{Remove, Update}, filename: "_testdata/project/assets/application.js.liquid", oldname: "_testdata/project/assets/application.js", op: watcher.Move},
	}

	for i, testcase := range testcases {
		w := createTestWatcher(t)
		w.Events = make(chan Event, len(testcases))
		info, _ := os.Stat(testcase.filename)
		w.onEvent(watcher.Event{Op: testcase.op, Path: testcase.filename, OldPath: testcase.oldname, FileInfo: info})
		assert.Equal(t, len(testcase.expectedOp), len(w.Events), fmt.Sprintf("testcase: %v", i))

		recievedEvents := OptSlice{}
		for i := 0; i < len(testcase.expectedOp); i++ {
			e := <-w.Events
			assert.Contains(t, testcase.filename, e.Path)
			recievedEvents = append(recievedEvents, e.Op)
		}

		sort.Sort(testcase.expectedOp)
		sort.Sort(recievedEvents)

		assert.Equal(t, testcase.expectedOp, recievedEvents, fmt.Sprintf("Wrong op in testcase %v", i))

		w.Stop()
	}
}

func TestFileWatcher_translateEvent(t *testing.T) {
	testcases := []struct {
		filename, oldname string
		op                watcher.Op
		expectedOp        OptSlice
	}{
		{expectedOp: OptSlice{Update}, filename: "_testdata/project/templates/template.liquid", op: watcher.Write},
		{expectedOp: OptSlice{Update}, filename: "_testdata/project/templates/customers/test.liquid", op: watcher.Create},
		{expectedOp: OptSlice{}, filename: "_testdata/project/config", op: watcher.Create},
		{expectedOp: OptSlice{Remove}, filename: "_testdata/project/templates/customers/test.liquid", op: watcher.Remove},
		{expectedOp: OptSlice{Remove, Update}, filename: "_testdata/project/assets/application.js.liquid", oldname: "_testdata/project/assets/application.js", op: watcher.Rename},
		{expectedOp: OptSlice{Remove, Update}, filename: "_testdata/project/assets/application.js.liquid", oldname: "_testdata/project/assets/application.js", op: watcher.Move},
	}

	for _, testcase := range testcases {
		info, _ := os.Stat(testcase.filename)
		evt := watcher.Event{
			Op:       testcase.op,
			Path:     testcase.filename,
			OldPath:  testcase.oldname,
			FileInfo: info,
		}

		w := createTestWatcher(t)
		events := w.translateEvent(evt)
		assert.Equal(t, len(testcase.expectedOp), len(events))
		recievedEvents := OptSlice{}
		for _, e := range events {
			recievedEvents = append(recievedEvents, e.Op)
		}
		sort.Sort(testcase.expectedOp)
		sort.Sort(recievedEvents)
		assert.Equal(t, testcase.expectedOp, recievedEvents)
		w.Stop()
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
		input, currentpath string
	}{
		{"_testdata/project/assets/app.js", "assets/app.js"},
		{"not/another/path/assets/app.js", "not/another/path/assets/app.js"},
	}
	w := createTestWatcher(t)
	for _, testcase := range testcases {
		assert.Equal(t, w.parsePath(testcase.input), testcase.currentpath)
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

func createTestWatcher(t *testing.T) *Watcher {
	e := &env.Env{
		Directory:    filepath.Join("_testdata", "project"),
		IgnoredFiles: []string{"config/"},
	}
	w, err := NewWatcher(e, filepath.Join("_testdata", "project", "config.yml"), map[string]string{})
	assert.Nil(t, err)
	return w
}

func createTestFilterHook(t *testing.T) watcher.FilterFileHookFunc {
	e := &env.Env{
		Directory:    filepath.Join("_testdata", "project"),
		IgnoredFiles: []string{"config/"},
	}
	hook, err := filterHook(e, filepath.Join("_testdata", "project", "config.yml"))
	assert.Nil(t, err)
	return hook
}
