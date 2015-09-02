package commands

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Shopify/themekit"
)

type BasicOptions struct {
	Client    themekit.ThemeClient
	Filenames []string
	EventLog  chan themekit.ThemeEvent
}

func (bo *BasicOptions) getEventLog() chan themekit.ThemeEvent {
	if bo.EventLog == nil {
		bo.EventLog = make(chan themekit.ThemeEvent)
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
		themekit.NotifyError(errors.New(errMsg))
	}
}

func extractStringSlice(key string, args map[string]interface{}) []string {
	var ok bool
	var strs []string
	if strs, ok = args[key].([]string); !ok {
		errMsg := fmt.Sprintf("%s is not of valid type", key)
		themekit.NotifyError(errors.New(errMsg))
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
		themekit.NotifyError(errors.New(errMsg))
	}
}

func extractBool(s *bool, key string, args map[string]interface{}) {
	var ok bool
	if args[key] == nil {
		return
	}

	if *s, ok = args[key].(bool); !ok {
		errMsg := fmt.Sprintf("%s is not of valid type", key)
		themekit.NotifyError(errors.New(errMsg))
	}
}

func extractThemeClient(t *themekit.ThemeClient, args map[string]interface{}) {
	var ok bool
	if args["themeClient"] == nil {
		return
	}

	if *t, ok = args["themeClient"].(themekit.ThemeClient); !ok {
		themekit.NotifyError(errors.New("themeClient is not of a valid type"))
	}
}

func extractThemeClients(args map[string]interface{}) []themekit.ThemeClient {
	if args["environments"] == nil {
		return []themekit.ThemeClient{}
	}

	var ok bool
	var environments themekit.Environments
	if environments, ok = args["environments"].(themekit.Environments); !ok {
		themekit.NotifyError(errors.New("environments is not of a valid type"))
	}
	clients := make([]themekit.ThemeClient, len(environments), len(environments))
	idx := 0
	for _, configuration := range environments {
		clients[idx] = themekit.NewThemeClient(configuration)
		idx++
	}
	return clients
}

func extractEventLog(el *chan themekit.ThemeEvent, args map[string]interface{}) {
	var ok bool

	if *el, ok = args["eventLog"].(chan themekit.ThemeEvent); !ok {
		themekit.NotifyError(errors.New("eventLog is not of a valid type"))
	}
}

func extractBasicOptions(args map[string]interface{}) BasicOptions {
	options := BasicOptions{}
	extractThemeClient(&options.Client, args)
	extractEventLog(&options.EventLog, args)
	options.Filenames = extractStringSlice("filenames", args)
	return options
}

func toClientAndFilesAsync(args map[string]interface{}, fn func(themekit.ThemeClient, []string) chan bool) chan bool {
	var ok bool
	var themeClient themekit.ThemeClient
	var filenames []string
	extractThemeClient(&themeClient, args)

	if filenames, ok = args["filenames"].([]string); !ok {
		themekit.NotifyError(errors.New("filenames are not of valid type"))
	}
	return fn(themeClient, filenames)
}

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
	etype     string `json:"type"`
}

func message(content string) themekit.ThemeEvent {
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
