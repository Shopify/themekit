package kit

import (
	"errors"
	"fmt"
	"log"
	"sync"
)

// ErrorReporter ... TODO
type ErrorReporter interface {
	Report(error)
}

type nullReporter struct{}

// ConsoleReporter ... TODO
type ConsoleReporter struct{}

// HaltExecutionReporter ... TODO
type HaltExecutionReporter struct{}

func (n nullReporter) Report(e error) {}

// Report ... TODO
func (c ConsoleReporter) Report(e error) {
	fmt.Println(RedText(e.Error()))
}

// Report ... TODO
func (h HaltExecutionReporter) Report(e error) {
	c := ConsoleReporter{}
	libraryInfo := fmt.Sprintf("%s%s%s", MessageSeparator, LibraryInfo(), MessageSeparator)
	c.Report(errors.New(libraryInfo))
	log.Fatal(e)
}

var reporter ErrorReporter = nullReporter{}
var errorQueue = make(chan error)
var mutex = &sync.Mutex{}

func synchronized(m *sync.Mutex, fn func()) {
	m.Lock()
	defer m.Unlock()
	fn()
}

// SetErrorReporter ... TODO
func SetErrorReporter(r ErrorReporter) {
	synchronized(mutex, func() {
		close(errorQueue)
		errorQueue = make(chan error)
	})

	reporter = r
	go func() {
		for {
			if err, ok := <-errorQueue; !ok {
				break
			} else {
				reporter.Report(err)
			}
		}
	}()
}

// NotifyErrorImmediately ... TODO
func NotifyErrorImmediately(err error) {
	synchronized(mutex, func() {
		reporter.Report(err)
	})
}

// NotifyError ... TODO
func NotifyError(err error) {
	synchronized(mutex, func() {
		go func() {
			errorQueue <- err
		}()
	})
}
