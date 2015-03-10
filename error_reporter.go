package phoenix

import "sync"

type ErrorReporter interface {
	Report(error)
}

var reporter ErrorReporter
var errorQueue chan error = make(chan error)
var mutex *sync.Mutex = &sync.Mutex{}

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

func notifyError(err error) {
	synchronized(mutex, func() {
		go func() {
			errorQueue <- err
		}()
	})
}
