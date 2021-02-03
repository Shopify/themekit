package file

import (
	"crypto/md5"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/Shopify/themekit/src/env"
	"github.com/radovskyb/watcher"
)

// Op describes the different types of file operations
type Op int

const (
	// Update is a file op where the file is updated
	Update Op = iota
	// Remove is a file op where the file is removed
	Remove
	// Skip is a file op where the remote file matches the local file so is not transferred
	Skip
	// Get is when a file should be re-fetched, used in download operations
	Get
)

var (
	// how long until we stop trying to drain events before emitting events
	drainTimeout = time.Second
	// the interval that the watcher polls the filesystem this needs to be less than
	// the drain timeout, otherwise debouncing will not work
	pollInterval = 500 * time.Millisecond
)

// Event decsribes a file change event
type Event struct {
	Op                Op
	Path              string
	LastKnownChecksum string
	checksum          string
}

// Watcher is the object used to watch files for change and notify on any events,
// these events can then be passed along to kit to be sent to shopify.
type Watcher struct {
	Events chan Event

	fsWatcher *watcher.Watcher
	directory string
	checksums map[string]string
}

// NewWatcher will create a new file change watching for a given directory defined
// in an environment
func NewWatcher(e *env.Env, configPath string, checksums map[string]string) (*Watcher, error) {
	fsWatcher := watcher.New()
	fsWatcher.IgnoreHiddenFiles(true)
	fsWatcher.FilterOps(watcher.Create, watcher.Write, watcher.Remove, watcher.Rename, watcher.Move)

	hook, err := filterHook(e, configPath)
	if err != nil {
		return nil, err
	}
	fsWatcher.AddFilterHook(hook)

	if err := fsWatcher.Add(e.Directory); err != nil {
		return nil, fmt.Errorf("Could not watch directory: %s", err)
	}
	for _, folder := range assetLocations {
		path := filepath.Join(e.Directory, folder)
		if err := fsWatcher.Add(path); err != nil && !os.IsNotExist(err) {
			return nil, fmt.Errorf("Could not watch directory %s: %s", path, err)
		}
	}

	return &Watcher{
		Events:    make(chan Event),
		directory: e.Directory,
		checksums: checksums,
		fsWatcher: fsWatcher,
	}, nil
}

func filterHook(e *env.Env, configPath string) (watcher.FilterFileHookFunc, error) {
	filter, err := NewFilter(e.Directory, e.IgnoredFiles, e.Ignores)
	if err != nil {
		return nil, err
	}
	return func(info os.FileInfo, fullPath string) error {
		if configPath != fullPath && filter.Match(fullPath) {
			return watcher.ErrSkip
		}
		return nil
	}, nil
}

// Watch will start the watcher actually receiving file change events and sending
// events to the Events channel
func (w *Watcher) Watch() {
	go w.watchFsEvents()
	go w.fsWatcher.Start(pollInterval)
}

func (w *Watcher) watchFsEvents() {
	for {
		select {
		case event, ok := <-w.fsWatcher.Event:
			if ok {
				w.onEvent(event)
			}
		case <-w.fsWatcher.Closed:
			w.Stop()
			return
		case <-w.fsWatcher.Error:
			// discard errors, they are not useful for users and the watcher deadlocks
			// if they are not read. The expected errors come from a directory being
			// deleted while being watched
		}
	}
}

func (w *Watcher) onEvent(event watcher.Event) bool {
	events := map[string]Event{}
	for _, e := range w.translateEvent(event) {
		events[e.Path] = e
	}
	if len(events) == 0 {
		return false
	}

	drainTimer := time.NewTimer(drainTimeout)
	defer drainTimer.Stop()
	for {
		select {
		case event, ok := <-w.fsWatcher.Event:
			if !ok {
				continue
			}
			for _, e := range w.translateEvent(event) {
				events[e.Path] = e
			}
			drainTimer.Reset(drainTimeout)
		case <-drainTimer.C:
			for _, e := range events {
				w.updateChecksum(e)
				w.Events <- e
			}
			return len(events) > 0
		}
	}
}

func (w *Watcher) updateChecksum(e Event) {
	if e.Op == Remove {
		delete(w.checksums, e.Path)
	} else if e.Op == Update {
		w.checksums[e.Path] = e.checksum
	}
}

func (w *Watcher) translateEvent(event watcher.Event) []Event {
	oldPath, currentPath := w.parsePath(event.OldPath), w.parsePath(event.Path)
	if event.IsDir() {
		if isEventType(event.Op, watcher.Create) {
			w.fsWatcher.Add(event.Path)
		}
	} else if isEventType(event.Op, watcher.Rename, watcher.Move) {
		return []Event{{Op: Remove, Path: oldPath}, {Op: Update, Path: currentPath, LastKnownChecksum: w.checksums[currentPath]}}
	} else if isEventType(event.Op, watcher.Remove) {
		return []Event{{Op: Remove, Path: currentPath}}
	} else if isEventType(event.Op, watcher.Create, watcher.Write) {
		checksum, err := fileChecksum(w.directory, currentPath)
		eventOp := Update
		if err == nil && checksum == w.checksums[currentPath] {
			eventOp = Skip
		}
		return []Event{{Op: eventOp, Path: currentPath, checksum: checksum, LastKnownChecksum: w.checksums[currentPath]}}
	}
	return []Event{}
}

func (w *Watcher) parsePath(path string) string {
	projectPath := pathToProject(w.directory, path)
	if projectPath == "" {
		return path
	}
	return projectPath
}

func isEventType(currentOp watcher.Op, allowedOps ...watcher.Op) bool {
	for _, op := range allowedOps {
		if currentOp == op {
			return true
		}
	}
	return false
}

// Stop will stop the Watcher from watching it's directories and clean
// up any go routines doing work.
func (w *Watcher) Stop() {
	w.fsWatcher.Close()
	for len(w.Events) > 0 { // drain events
		<-w.Events
	}
}

func fileChecksum(dir, src string) (string, error) {
	sum := md5.New()
	s, err := os.Open(filepath.Join(dir, src))
	if err != nil {
		return "", err
	}
	defer s.Close()
	_, err = io.Copy(sum, s)
	return fmt.Sprintf("%x", sum.Sum(nil)), err
}
