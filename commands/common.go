package commands

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/csaunders/phoenix"
)

type BasicOptions struct {
	Client    phoenix.ThemeClient
	Filenames []string
	EventLog  chan phoenix.ThemeEvent
}

func (bo *BasicOptions) getEventLog() chan phoenix.ThemeEvent {
	if bo.EventLog == nil {
		bo.EventLog = make(chan phoenix.ThemeEvent)
	}
	return bo.EventLog
}

func extractString(s *string, key string, args map[string]interface{}) {
	var ok bool
	if args[key] == nil {
		return
	}

	if *s, ok = args[key].(string); !ok {
		errMsg := fmt.Sprintf("%s is not of valid type", key)
		phoenix.NotifyError(errors.New(errMsg))
	}
}

func extractStringSlice(key string, args map[string]interface{}) []string {
	var ok bool
	var strs []string
	if strs, ok = args[key].([]string); !ok {
		errMsg := fmt.Sprintf("%s is not of valid type", key)
		phoenix.NotifyError(errors.New(errMsg))
	}
	return strs
}

func extractInt(s *int, key string, args map[string]interface{}) {
	var ok bool
	if args[key] == nil {
		return
	}

	if *s, ok = args[key].(int); !ok {
		errMsg := fmt.Sprintf("%s is not of valid type", key)
		phoenix.NotifyError(errors.New(errMsg))
	}
}

func extractBool(s *bool, key string, args map[string]interface{}) {
	var ok bool
	if args[key] == nil {
		return
	}

	if *s, ok = args[key].(bool); !ok {
		errMsg := fmt.Sprintf("%s is not of valid type", key)
		phoenix.NotifyError(errors.New(errMsg))
	}
}

func extractThemeClient(t *phoenix.ThemeClient, args map[string]interface{}) {
	var ok bool
	if args["themeClient"] == nil {
		return
	}

	if *t, ok = args["themeClient"].(phoenix.ThemeClient); !ok {
		phoenix.NotifyError(errors.New("themeClient is not of a valid type"))
	}
}

func extractEventLog(el *chan phoenix.ThemeEvent, args map[string]interface{}) {
	var ok bool

	if *el, ok = args["eventLog"].(chan phoenix.ThemeEvent); !ok {
		phoenix.NotifyError(errors.New("eventLog is not of a valid type"))
	}
}

func extractBasicOptions(args map[string]interface{}) BasicOptions {
	options := BasicOptions{}
	extractThemeClient(&options.Client, args)
	extractEventLog(&options.EventLog, args)
	options.Filenames = extractStringSlice("filenames", args)
	return options
}

func toClientAndFilesAsync(args map[string]interface{}, fn func(phoenix.ThemeClient, []string) chan bool) chan bool {
	var ok bool
	var themeClient phoenix.ThemeClient
	var filenames []string
	extractThemeClient(&themeClient, args)

	if filenames, ok = args["filenames"].([]string); !ok {
		phoenix.NotifyError(errors.New("filenames are not of valid type"))
	}
	return fn(themeClient, filenames)
}

func drainErrors(errs chan error) {
	for {
		if err := <-errs; err != nil {
			phoenix.NotifyError(err)
		} else {
			break
		}
	}
}

func mergeEvents(dest chan phoenix.ThemeEvent, chans []chan phoenix.ThemeEvent) {
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

func logEvent(event phoenix.ThemeEvent, eventLog chan phoenix.ThemeEvent) {
	go func() {
		eventLog <- event
	}()
}

type basicEvent struct {
	Formatter func(b basicEvent) string
	EventType string `json:"event_type"`
	Target    string `json:"target"`
	Title     string `json:"title"`
	etype     string `json:"type"`
}

func message(content string) phoenix.ThemeEvent {
	return basicEvent{
		Formatter: func(b basicEvent) string { return content },
		EventType: "message",
		Title:     "Notice",
		etype:     "basicEvent",
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
