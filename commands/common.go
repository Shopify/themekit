package commands

import (
	"encoding/json"

	"github.com/Shopify/themekit"
)

func drainErrors(errs chan error) {
	for {
		if err := <-errs; err != nil {
			themekit.NotifyError(err)
		} else {
			break
		}
	}
}

func mergeEvents(dest chan themekit.ThemeEvent, chans []chan themekit.ThemeEvent) {
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

func logEvent(event themekit.ThemeEvent, eventLog chan themekit.ThemeEvent) {
	go func() {
		eventLog <- event
	}()
}

type basicEvent struct {
	Formatter func(b basicEvent) string
	EventType string `json:"event_type"`
	Target    string `json:"target"`
	Title     string `json:"title"`
	Etype     string `json:"type"`
}

func message(content string) themekit.ThemeEvent {
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
