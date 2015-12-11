package themekit

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/Shopify/themekit/theme"
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

type NoOpEvent struct {
}

func (e NoOpEvent) String() string {
	return ""
}

func (e NoOpEvent) Successful() bool {
	return false
}

func (e NoOpEvent) Error() error {
	return nil
}

func (e NoOpEvent) AsJSON() ([]byte, error) {
	return []byte{}, errors.New("cannot encode NoOpEvents")
}

type APIAssetEvent struct {
	Host      string `json:"host"`
	AssetKey  string `json:"asset_key"`
	EventType string `json:"event_type"`
	Code      int    `json:"status_code"`
	err       error  `json:"error,omitempty"` // TODO: err is unexported; json binding is not going to work
	etype     string `json:"type"`            // TODO: same here, unexported, no json binding
}

func NewAPIAssetEvent(r *http.Response, e AssetEvent, err error) APIAssetEvent {
	event := APIAssetEvent{
		AssetKey:  e.Asset().Key,
		EventType: e.Type().String(),
		etype:     "APIAssetEvent",
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
		return fmt.Sprintf(
			"Successfully performed %s operation for file %s to %s",
			GreenText(a.EventType),
			BlueText(a.AssetKey),
			YellowText(a.Host),
		)
	} else if a.Code == 422 {
		return RedText(fmt.Sprintf("Could not upload %s:\n\t%s", a.AssetKey, a.err))
	} else if a.Code == 403 {
		return fmt.Sprintf(
			"[%s]Cannot remove files that would make a theme invalid. %s",
			RedText(fmt.Sprintf("%d", a.Code)),
			BlueText(a.AssetKey),
		)
	} else if a.Code == 404 {
		return fmt.Sprintf(
			"[%s]Could not complete operation because %s does not exist",
			RedText(fmt.Sprintf("%d", a.Code)),
			BlueText(a.AssetKey),
		)
	} else {
		return fmt.Sprintf(
			"[%s]Could not perform %s to %s at %s\n\t%s",
			RedText(fmt.Sprintf("%d", a.Code)),
			YellowText(a.EventType),
			BlueText(a.AssetKey),
			YellowText(a.Host),
			a.err,
		)
	}
}

func (a APIAssetEvent) Successful() bool {
	return a.Code >= 200 && a.Code < 300
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
	etype       string `json:"type"`
}

func NewAPIThemeEvent(r *http.Response, err error) APIThemeEvent {
	if err != nil {
		return APIThemeEvent{Host: "Unknown Host", Code: ThemeEventErrorCode, err: err, etype: "APIThemeEvent"}
	}
	event := APIThemeEvent{Host: r.Request.URL.Host, Code: r.StatusCode, etype: "APIThemeEvent"}

	if event.Successful() {
		populateThemeData(&event, r)
	} else {
		populateAPIErrorData(&event, r)
	}
	return event
}

func (t APIThemeEvent) String() string {
	if t.Successful() {
		return fmt.Sprintf(
			"[%s]Modifications made to theme '%s' with id of %s on shop %s",
			GreenText(fmt.Sprintf("%d", t.Code)),
			BlueText(t.ThemeName),
			BlueText(fmt.Sprintf("%d", t.ThemeId)),
			YellowText(t.Host),
		)
	} else {
		return fmt.Sprintf(
			"[%s]Encoutered error with request to %s\n\t%s",
			RedText(fmt.Sprintf("%d", t.Code)),
			YellowText(t.Host),
			t.err,
		)
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
	var container map[string]theme.Theme
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
