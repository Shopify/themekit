package kit

import (
	"fmt"
	"net/http"
)

type kitError interface {
	Error() string
	Fatal() bool
}

// ThemeError is an error that is encountered during a request on a theme.
type KitError struct {
	err error
}

func (err KitError) Fatal() bool {
	return true
}

func (err KitError) Error() string {
	return err.err.Error()
}

// ThemeError is an error that is encountered during a request on a theme.
type ThemeError struct {
	resp shopifyResponse
}

func (err ThemeError) Fatal() bool {
	return !err.resp.Successful()
}

func (err ThemeError) Error() string {
	return fmt.Sprintf(`[%s]Theme request encountered status at host <%s>
	Status text: %s
	Errors: %s`,
		RedText(fmt.Sprintf("%d", err.resp.Code)),
		YellowText(err.resp.Host),
		RedText(http.StatusText(err.resp.Code)),
		err.resp.fmtErrors(),
	)
}

// AssetError is an error that is encountered during a request on a single asset.
type AssetError struct {
	resp shopifyResponse
}

func (err AssetError) Fatal() bool {
	return err.resp.Code != 404 && err.resp.Code >= 400
}

func (err AssetError) Error() string {
	return fmt.Sprintf(`[%s]Asset Perform %s to %s at host <%s>
	Status text: %s
	Errors: %s`,
		RedText(fmt.Sprintf("%d", err.resp.Code)),
		YellowText(err.resp.EventType),
		BlueText(err.resp.Asset.Key),
		YellowText(err.resp.Host),
		RedText(http.StatusText(err.resp.Code)),
		err.resp.fmtErrors(),
	)
}

// List error are errors that are encountered during a request of remote assets
type ListError struct {
	resp shopifyResponse
}

func (err ListError) Fatal() bool {
	return err.resp.Code >= 400
}

func (err ListError) Error() string {
	return fmt.Sprintf(`[%s]Assets Perform %s at host <%s>
	Status text: %s
	Errors: %s`,
		RedText(fmt.Sprintf("%d", err.resp.Code)),
		YellowText(err.resp.EventType),
		YellowText(err.resp.Host),
		RedText(http.StatusText(err.resp.Code)),
		err.resp.fmtErrors(),
	)
}
