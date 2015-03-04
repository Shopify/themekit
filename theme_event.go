package phoenix

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

// Using a 500 Error Code is disingenuous since
// it could be an error internal to the library
// and not necessarily with the Service Provider
const ThemeEventErrorCode int = 999

type ThemeEvent interface {
	String() string
	Successful() bool
	Error() error
	AsJSON() ([]byte, error)
}

type APIAssetEvent struct {
	Host      string `json:"host"`
	AssetKey  string `json:"asset_key"`
	EventType string `json:"event_type"`
	Code      int    `json:"status_code"`
	err       error  `json:"error,omitempty"`
}

func NewAPIAssetEvent(r *http.Response, e AssetEvent, err error) APIAssetEvent {
	event := APIAssetEvent{
		AssetKey:  e.Asset().Key,
		EventType: e.Type().String(),
	}
	if err != nil {
		event.Host = "Host Unknown"
		event.Code = ThemeEventErrorCode
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

func (a APIAssetEvent) String() string {
	if a.Successful() {
		return fmt.Sprintf("Successfully performed %s operation for file %s to %s", a.EventType, BlueText(a.AssetKey), BlueText(a.Host))
	} else if a.Code == 422 {
		return fmt.Sprintf("Could not upload %s:\n\t%s", a.AssetKey, a.err)
	} else {
		return fmt.Sprintf("[%d]Could not perform %s to %s at %s\n\t%s", a.Code, YellowText(a.EventType), BlueText(a.AssetKey), BlueText(a.Host), a.err)
	}
}

func (a APIAssetEvent) Successful() bool {
	return a.Code >= 200 && a.Code <= 300
}

func (a APIAssetEvent) Error() error {
	return a.err
}

func (a APIAssetEvent) AsJSON() ([]byte, error) {
	return json.Marshal(a)
}

type AssetError struct {
	Messages []string `json:"asset"`
}

type APIThemeEvent struct {
	Host        string `json:"host"`
	ThemeName   string `json:"name"`
	ThemeId     int64  `json:"theme_id"`
	Code        int    `json:"status_code"`
	Previewable bool   `json:"previewable,omitempty"`
	err         error  `json:"error,omitempty"`
}

func NewAPIThemeEvent(r *http.Response, err error) APIThemeEvent {
	if err != nil {
		return APIThemeEvent{Host: "Unknown Host", Code: ThemeEventErrorCode, err: err}
	}
	event := APIThemeEvent{Host: r.Request.URL.Host, Code: r.StatusCode}

	if event.Successful() {
		populateThemeData(&event, r)
	} else {
		populateAPIErrorData(&event, r)
	}
	return event
}

func (t APIThemeEvent) String() string {
	if t.Successful() {
		return fmt.Sprintf("[%d]Theme called '%s' with id of %d for shop %s", t.Code, t.ThemeName, t.ThemeId, t.Host)
	} else {
		return fmt.Sprintf("[%d]Encoutered error with request to %s\n\t%s", t.Code, t.Host, t.err)
	}
}

func (t APIThemeEvent) Successful() bool {
	return t.Code >= 200 && t.Code < 300
}

func (t APIThemeEvent) Error() error {
	return t.err
}

func (t APIThemeEvent) AsJSON() ([]byte, error) {
	return json.Marshal(t)
}

func (t *APIThemeEvent) markIfHasError(err error) bool {
	if err != nil {
		t.Code = ThemeEventErrorCode
		t.err = err
		return true
	}
	return false
}

func populateThemeData(e *APIThemeEvent, r *http.Response) {
	var container map[string]Theme
	bytes, err := ioutil.ReadAll(r.Body)
	if e.markIfHasError(err) {
		return
	}
	if e.markIfHasError(json.Unmarshal(bytes, &container)) {
		return
	}
	theme := container["theme"]
	e.ThemeId = theme.Id
	e.ThemeName = theme.Name
	e.Previewable = theme.Previewable
}

func populateAPIErrorData(e *APIThemeEvent, r *http.Response) {
	data, err := ioutil.ReadAll(r.Body)
	if !e.markIfHasError(err) {
		e.err = errors.New(string(data))
	}
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
