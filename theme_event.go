package phoenix

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

type ThemeEvent interface {
	String() string
	Successful() bool
	Error() error
	AsJSON() ([]byte, error)
}

type APIThemeEvent struct {
	Host      string `json:"host"`
	AssetKey  string `json:"asset_key"`
	EventType string `json:"event_type"`
	Code      int    `json:"status_code"`
	err       error  `json:"error,omitempty"`
}

func NewAPIThemeEvent(r *http.Response, e AssetEvent, err error) APIThemeEvent {
	event := APIThemeEvent{
		AssetKey:  e.Asset().Key,
		EventType: e.Type().String(),
	}
	if err != nil {
		event.Host = "Host Unknown"
		event.Code = 500
		event.err = err
	} else {
		event.Host = r.Request.URL.Host
		event.Code = r.StatusCode
		if !event.Successful() {
			event.err = extractAssetAPIErrors(ioutil.ReadAll(r.Body))
		}
	}

	return event
}

func (a APIThemeEvent) String() string {
	if a.Successful() {
		return fmt.Sprintf("Successfully performed %s operation for file %s to %s", a.EventType, BlueText(a.AssetKey), BlueText(a.Host))
	} else if a.Code == 422 {
		return fmt.Sprintf("Could not upload %s:\n\t%s", a.AssetKey, a.err)
	} else {
		return fmt.Sprintf("[%d]Could not perform %s to %s at %s", a.Code, YellowText(a.EventType), BlueText(a.AssetKey), BlueText(a.Host))
	}
}

func (a APIThemeEvent) Successful() bool {
	return a.Code >= 200 && a.Code <= 300
}

func (a APIThemeEvent) Error() error {
	return a.err
}

func (a APIThemeEvent) AsJSON() ([]byte, error) {
	return json.Marshal(a)
}

type AssetError struct {
	Messages []string `json:"asset"`
}

func extractAssetAPIErrors(data []byte, err error) error {
	if err != nil {
		return err
	}

	var assetErrors map[string]AssetError
	err = json.Unmarshal(data, &assetErrors)

	if err != nil {
		return err
	}
	return errors.New(strings.Join(assetErrors["errors"].Messages, "\n"))
}
