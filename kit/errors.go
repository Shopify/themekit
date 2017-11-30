package kit

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/fatih/color"
)

var (
	red    = color.New(color.FgRed).SprintFunc()
	yellow = color.New(color.FgYellow).SprintFunc()
	blue   = color.New(color.FgBlue).SprintFunc()
	green  = color.New(color.FgGreen).SprintFunc()
)

// Error is an error that can be determined if it is fatal to the applications
// operation.
type Error interface {
	Error() string
	Fatal() bool
}

type kitError struct {
	err error
}

func (err kitError) Fatal() bool {
	return true
}

func (err kitError) Error() string {
	return err.err.Error()
}

type themeError struct {
	resp       ShopifyResponse
	requestErr requestError
}

func newThemeError(resp ShopifyResponse) themeError {
	return themeError{resp: resp, requestErr: resp.Errors}
}

func (err themeError) Fatal() bool {
	return !err.resp.Successful()
}

func (err themeError) Error() string {
	return fmt.Sprintf(`Theme request encountered status at host %s
	Status: %s %s
	Errors: %s`,
		yellow(err.resp.Host),
		red(err.resp.Code),
		red(http.StatusText(err.resp.Code)),
		yellow(err.requestErr),
	)
}

type assetError struct {
	resp       ShopifyResponse
	requestErr requestError
}

func newAssetError(resp ShopifyResponse) assetError {
	err := assetError{resp: resp, requestErr: resp.Errors}
	err.generateHints()
	return err
}

func (err assetError) Fatal() bool {
	return err.resp.Code != 404 &&
		(err.resp.Code < 200 || err.resp.Code >= 400 || err.requestErr.Any())
}

func (err *assetError) generateHints() {
	if err.resp.EventType == Remove && err.resp.Code == 403 {
		err.requestErr.AddS("This file is critical and removing it would cause your theme to become non-functional.")
	}
	if err.resp.EventType == Update && err.resp.Code == 404 {
		err.requestErr.AddS("This file is not part of your theme.")
	}
	if err.resp.EventType == Update && err.resp.Code == 409 {
		err.requestErr.AddS(`
There have been changes to this file made remotely.

You can solve this by running 'theme download' to get the most recent copy of this file.
Running 'theme download' will overwrite any changes you have made so make sure to make
a backup.

If you are certain that you want to overwrite any changes then use the --force flag
		`)
	}
}

func (err assetError) Error() string {
	return fmt.Sprintf(`Asset Perform %s to %s at host %s
	Status: %s %s
	Errors: %s`,
		yellow(err.resp.EventType),
		blue(err.resp.Asset.Key),
		yellow(err.resp.Host),
		red(err.resp.Code),
		red(http.StatusText(err.resp.Code)),
		yellow(err.requestErr),
	)
}

type listError struct {
	resp       ShopifyResponse
	requestErr requestError
}

func newListError(resp ShopifyResponse) listError {
	return listError{resp: resp, requestErr: resp.Errors}
}

func (err listError) Fatal() bool {
	return err.resp.Code < 200 || err.resp.Code >= 400 || err.requestErr.Any()
}

func (err listError) Error() string {
	return fmt.Sprintf(`Assets Perform %s at host <%s>
	Status: %s %s
	Errors: %s`,
		yellow(err.resp.EventType),
		yellow(err.resp.Host),
		red(err.resp.Code),
		red(http.StatusText(err.resp.Code)),
		yellow(err.requestErr),
	)
}

type generalRequestError struct {
	Error string `json:"errors"`
}

type requestError struct {
	Syntax []string `json:"asset"`
	Theme  []string `json:"src"`
	Other  []string `json:"-"`
}

func (err requestError) all() []string {
	return append(err.Syntax, append(err.Theme, err.Other...)...)
}

func (err requestError) Any() bool {
	return len(err.all()) > 0
}

func (err requestError) Error() string {
	return err.String()
}

func (err requestError) String() string {
	if err.Any() {
		return strings.Join(err.all(), "\n\t\t")
	}
	return "none"
}

func (err *requestError) Add(other generalRequestError) {
	if other.Error != "" {
		err.Other = append(err.Other, other.Error)
	}
}

func (err *requestError) AddS(other string) {
	if other != "" {
		err.Other = append(err.Other, other)
	}
}

func (err *requestError) AddE(other error) {
	if other != nil {
		err.Other = append(err.Other, other.Error())
	}
}
