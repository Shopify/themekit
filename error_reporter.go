package themekit

import (
	"errors"
	"fmt"
	"log"
	"sync"
)

type ErrorReporter interface {
	Report(error)
}

type nullReporter struct{}
type ConsoleReporter struct{}
type HaltExecutionReporter struct{}

func (n nullReporter) Report(e error) {}

func (c ConsoleReporter) Report(e error) {
	fmt.Println(RedText(e.Error()))
}

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

func SetErrorReporter(r ErrorReporter) {
	synchronized(mutex, func() {
		close(errorQueue)
		errorQueue = make(chan error)
	})

	reporter = r
	go func() {
		for {
			if err, ok := <-errorQueue; !ok {
				return
			} else {
				reporter.Report(err)
			}
		}
	}()
}

func NotifyErrorImmediately(err error) {
	synchronized(mutex, func() {
		reporter.Report(err)
	})
}

func NotifyError(err error) {
	synchronized(mutex, func() {
		go func() {
			errorQueue <- err
		}()
	})
}
