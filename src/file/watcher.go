package file

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/Shopify/themekit/src/env"
	"github.com/fsnotify/fsnotify"
)

// Op describes the different types of file operations
type Op int

const (
	// Update is a file op where the file is updated
	Update Op = iota
	// Remove is a file op where the file is removed
	Remove
)

// Event decsribes a file change event
type Event struct {
	Op   Op
	Path string
}

type debouncer func(timeout time.Duration, events, complete chan fsnotify.Event)

// Watcher is the object used to watch files for change and notify on any events,
// these events can then be passed along to kit to be sent to shopify.
type Watcher struct {
	fsWatcher       *fsnotify.Watcher
	filter          Filter
	notify          string
	directory       string
	configPath      string
	events          chan Event
	debounceTimeout time.Duration
	idleTimeout     time.Duration
}

// NewWatcher will create a new file change watching for a a given directory defined
// in an environment
func NewWatcher(e *env.Env, configPath string) (*Watcher, error) {
	filter, err := NewFilter(e.Directory, e.IgnoredFiles, e.Ignores)
	if err != nil {
		return nil, err
	}

	return &Watcher{
		configPath:      configPath,
		directory:       e.Directory,
		filter:          filter,
		notify:          e.Notify,
		debounceTimeout: 1100 * time.Millisecond,
		idleTimeout:     time.Second,
	}, nil
}

// Watch will start the watcher actually receiving file change events and sending
// events to the Events channel
func (w *Watcher) Watch() (chan Event, error) {
	fsWatcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	w.fsWatcher = fsWatcher
	w.events = make(chan Event)

	go w.watchFsEvents(fsWatcher.Events, debounce)

	if err := fsWatcher.Add(w.configPath); err != nil {
		return nil, fmt.Errorf("Could not config path: %s", err)
	}

	return w.events, filepath.Walk(w.directory, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() && !w.filter.Match(path) && path != w.directory {
			if err := fsWatcher.Add(path); err != nil {
				return fmt.Errorf("Could not watch directory %s: %s", path, err)
			}
		}
		return nil
	})
}

func (w *Watcher) watchFsEvents(events chan fsnotify.Event, debounce debouncer) {
	fileEvents := make(map[string]chan fsnotify.Event)
	complete := make(chan fsnotify.Event)
	idleTimer := time.NewTimer(w.idleTimeout)
	defer idleTimer.Stop()

	for {
		select {
		case event := <-complete:
			projectPath := pathToProject(w.directory, event.Name)
			if projectPath == "" {
				projectPath = event.Name
			}
			e := Event{Op: Update, Path: projectPath}
			if event.Op&fsnotify.Remove == fsnotify.Remove || event.Op&fsnotify.Rename == fsnotify.Rename {
				e.Op = Remove
			}
			w.events <- e
			delete(fileEvents, event.Name)
			if len(fileEvents) == 0 {
				idleTimer.Reset(w.idleTimeout)
			}
		case event, more := <-events:
			if !more {
				close(w.events)
				return
			}

			if event.Op == fsnotify.Chmod || (w.configPath != event.Name && w.filter.Match(event.Name)) {
				continue
			}

			idleTimer.Stop()
			if _, ok := fileEvents[event.Name]; !ok {
				fileEvents[event.Name] = make(chan fsnotify.Event)
				go debounce(w.debounceTimeout, fileEvents[event.Name], complete)
			}
			fileEvents[event.Name] <- event
		case <-idleTimer.C:
			w.onIdle()
		}
	}
}

// Stop will stop the Watcher from watching it's directories and clean
// up any go routines doing work.
func (w *Watcher) Stop() {
	if w.fsWatcher == nil {
		return
	}
	w.fsWatcher.Close()
}

func (w *Watcher) onIdle() {
	if w.notify == "" {
		return
	}
	os.Create(w.notify)
	os.Chtimes(w.notify, time.Now(), time.Now())
}

func debounce(timeout time.Duration, incoming, complete chan fsnotify.Event) {
	var event fsnotify.Event
	for {
		select {
		case event = <-incoming:
		case <-time.After(timeout):
			complete <- event
			return
		}
	}
}
