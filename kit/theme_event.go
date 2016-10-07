package kit

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/Shopify/themekit/theme"
)

// ThemeEventErrorCode ... Using a 500 Error Code is disingenuous since
// it could be an error internal to the library
// and not necessarily with the Service Provider
const ThemeEventErrorCode int = 999

// ThemeEvent is an interface that describes all the events that pass through the
// Theme clients event log
type ThemeEvent interface {
	String() string
	Successful() bool
	Error() error
	AsJSON() ([]byte, error)
}

// AssetEvent is an interface that describes events that are related to assets that
// are processed through the eventlog
type AssetEvent interface {
	Asset() theme.Asset
	Type() EventType
}

type apiAssetEvent struct {
	Host      string `json:"host"`
	AssetKey  string `json:"asset_key"`
	EventType string `json:"event_type"`
	Code      int    `json:"status_code"`
	err       error
	etype     string
}

func newAPIAssetEvent(r *http.Response, e AssetEvent, err error) apiAssetEvent {
	event := apiAssetEvent{
		AssetKey:  e.Asset().Key,
		EventType: e.Type().String(),
		etype:     "apiAssetEvent",
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

func (a apiAssetEvent) String() string {
	if a.Successful() {
		return fmt.Sprintf(
			"Successfully performed %s operation for file %s to %s",
			GreenText(a.EventType),
			BlueText(a.AssetKey),
			YellowText(a.Host),
		)
	} else if a.Code == 422 {
		return RedText(fmt.Sprintf("Could not upload %s:\n\t%s", a.AssetKey, a.err))
	} else if a.Code == 403 || a.Code == 401 {
		return fmt.Sprintf(
			"[%s]Insufficient permissions to perform %s to %s",
			RedText(fmt.Sprintf("%d", a.Code)),
			YellowText(a.EventType),
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

func (a apiAssetEvent) Successful() bool {
	return a.Code >= 200 && a.Code < 300
}

func (a apiAssetEvent) Error() error {
	return a.err
}

func (a apiAssetEvent) AsJSON() ([]byte, error) {
	return json.Marshal(a)
}

type assetError struct {
	Messages []string `json:"asset"`
}

type apiThemeEvent struct {
	Host        string `json:"host"`
	ThemeName   string `json:"name"`
	ThemeID     int64  `json:"theme_id"`
	Code        int    `json:"status_code"`
	Previewable bool   `json:"previewable,omitempty"`
	err         error
	etype       string
}

func newAPIThemeEvent(r *http.Response, err error) apiThemeEvent {
	if err != nil {
		return apiThemeEvent{Host: "Unknown Host", Code: ThemeEventErrorCode, err: err, etype: "APIThemeEvent"}
	}
	event := apiThemeEvent{Host: r.Request.URL.Host, Code: r.StatusCode, etype: "APIThemeEvent"}

	if event.Successful() {
		populateThemeData(&event, r)
	} else {
		populateAPIErrorData(&event, r)
	}
	return event
}

func (t apiThemeEvent) String() string {
	if t.Successful() {
		return fmt.Sprintf(
			"[%s]Modifications made to theme '%s' with id of %s on shop %s",
			GreenText(fmt.Sprintf("%d", t.Code)),
			BlueText(t.ThemeName),
			BlueText(fmt.Sprintf("%d", t.ThemeID)),
			YellowText(t.Host),
		)
	}

	return fmt.Sprintf(
		"[%s]Encoutered error with request to %s\n\t%s",
		RedText(fmt.Sprintf("%d", t.Code)),
		YellowText(t.Host),
		t.err,
	)
}

func (t apiThemeEvent) Successful() bool {
	return t.Code >= 200 && t.Code < 300
}

func (t apiThemeEvent) Error() error {
	return t.err
}

func (t apiThemeEvent) AsJSON() ([]byte, error) {
	return json.Marshal(t)
}

func (t *apiThemeEvent) markIfHasError(err error) bool {
	if err != nil {
		t.Code = ThemeEventErrorCode
		t.err = err
		return true
	}
	return false
}

func populateThemeData(e *apiThemeEvent, r *http.Response) {
	var container map[string]theme.Theme
	bytes, err := ioutil.ReadAll(r.Body)
	if e.markIfHasError(err) {
		return
	}
	if e.markIfHasError(json.Unmarshal(bytes, &container)) {
		return
	}
	theme := container["theme"]
	e.ThemeID = theme.ID
	e.ThemeName = theme.Name
	e.Previewable = theme.Previewable
}

func populateAPIErrorData(e *apiThemeEvent, r *http.Response) {
	data, err := ioutil.ReadAll(r.Body)
	if !e.markIfHasError(err) {
		e.err = errors.New(string(data))
	}
}

func extractAssetAPIErrors(data []byte, err error) error {
	if err != nil {
		return err
	}

	var assetErrors map[string]assetError
	err = json.Unmarshal(data, &assetErrors)

	if err != nil {
		return err
	}
	return errors.New(strings.Join(assetErrors["errors"].Messages, "\n"))
}

type basicEvent struct {
	Formatter func(b basicEvent) string
	EventType string `json:"event_type"`
	Target    string `json:"target"`
	Title     string `json:"title"`
	Etype     string `json:"type"`
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
