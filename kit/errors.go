package kit

import (
	"fmt"
	"net/http"
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
	resp ShopifyResponse
}

func (err themeError) Fatal() bool {
	return !err.resp.Successful()
}

func (err themeError) Error() string {
	return fmt.Sprintf(`[%s]Theme request encountered status at host <%s>
	Status text: %s
	Errors: %s`,
		RedText(err.resp.Code),
		YellowText(err.resp.Host),
		RedText(http.StatusText(err.resp.Code)),
		err.resp.Errors,
	)
}

type assetError struct {
	resp ShopifyResponse
}

func (err assetError) Fatal() bool {
	return err.resp.Code != 404 && err.resp.Code >= 400
}

func (err assetError) Error() string {
	return fmt.Sprintf(`[%s]Asset Perform %s to %s at host <%s>
	Status text: %s
	Errors: %s`,
		RedText(err.resp.Code),
		YellowText(err.resp.EventType),
		BlueText(err.resp.Asset.Key),
		YellowText(err.resp.Host),
		RedText(http.StatusText(err.resp.Code)),
		err.resp.Errors,
	)
}

type listError struct {
	resp ShopifyResponse
}

func (err listError) Fatal() bool {
	return err.resp.Code >= 400
}

func (err listError) Error() string {
	return fmt.Sprintf(`[%s]Assets Perform %s at host <%s>
	Status text: %s
	Errors: %s`,
		RedText(err.resp.Code),
		YellowText(err.resp.EventType),
		YellowText(err.resp.Host),
		RedText(http.StatusText(err.resp.Code)),
		err.resp.Errors,
	)
}
