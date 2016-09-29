package commands

import (
	"encoding/json"

	"github.com/Shopify/themekit/kit"
)

func drainErrors(errs chan error) {
	for {
		if err := <-errs; err != nil {
			kit.NotifyError(err)
		} else {
			break
		}
	}
}

func mergeEvents(dest chan kit.ThemeEvent, chans []chan kit.ThemeEvent) {
	go func() {
		for _, ch := range chans {
			var ok = true
			for ok {
				if ev, ok := <-ch; ok {
					dest <- ev
				}
			}
			close(ch)
		}
	}()
}

func logEvent(event kit.ThemeEvent, eventLog chan kit.ThemeEvent) {
	go func() {
		eventLog <- event
	}()
}

func prepareChannel(args Args) (rawEvents, throttledEvents chan kit.AssetEvent) {
	rawEvents = make(chan kit.AssetEvent)
	if args.Bucket == nil {
		return rawEvents, rawEvents
	}

	foreman := kit.NewForeman(args.Bucket)
	foreman.JobQueue = rawEvents
	foreman.WorkerQueue = make(chan kit.AssetEvent)
	foreman.IssueWork()
	return foreman.JobQueue, foreman.WorkerQueue
}

type basicEvent struct {
	Formatter func(b basicEvent) string
	EventType string `json:"event_type"`
	Target    string `json:"target"`
	Title     string `json:"title"`
	Etype     string `json:"type"`
}

func message(content string) kit.ThemeEvent {
	return basicEvent{
		Formatter: func(b basicEvent) string { return content },
		EventType: "message",
		Title:     "Notice",
		Etype:     "basicEvent",
	}
}

func (b basicEvent) String() string {
	return b.Formatter(b)
}

func (b basicEvent) Successful() bool {
	return true
}

func (b basicEvent) Error() error {
	return nil
}

func (b basicEvent) AsJSON() ([]byte, error) {
	return json.Marshal(b)
}
