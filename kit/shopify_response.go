package kit

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
)

// ShopifyResponse is a general response for all server requests. It will format
// errors from any bad responses from the server. If the response is Successful()
// then the data item that you requested should be defined. If it was a theme request
// then Theme will be defined. If you have mad an asset query then Assets will be
// defined. If you did an action on a single asset then Asset will be defined.
type ShopifyResponse struct {
	Type      requestType  `json:"-"`
	Host      string       `json:"host"`
	URL       *url.URL     `json:"url"`
	Code      int          `json:"status_code"`
	Theme     Theme        `json:"theme"`
	Asset     Asset        `json:"asset"`
	Assets    []Asset      `json:"assets"`
	EventType EventType    `json:"event_type"`
	Errors    requestError `json:"errors"`
}

func newShopifyResponse(rtype requestType, event EventType, resp *http.Response, err error) (*ShopifyResponse, Error) {
	if resp == nil || err != nil {
		return nil, kitError{err}
	}
	defer resp.Body.Close()

	newResponse := &ShopifyResponse{
		Type:      rtype,
		Host:      resp.Request.URL.Host,
		URL:       resp.Request.URL,
		Code:      resp.StatusCode,
		EventType: event,
	}

	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, kitError{err}
	}

	err = json.Unmarshal(bytes, &newResponse)
	if err != nil {
		reqErr := generalRequestError{}
		json.Unmarshal(bytes, &reqErr)
		newResponse.Errors.Add(reqErr)
	}

	return newResponse, newResponse.Error()
}

// Successful will return true if the response code >= 200 and < 300 and if no
// errors were returned from the server.
func (resp ShopifyResponse) Successful() bool {
	return resp.Code >= 200 && resp.Code < 300 && !resp.Errors.Any()
}

func (resp ShopifyResponse) isThemeRequest() bool {
	return resp.Type == themeRequest
}

func (resp ShopifyResponse) isAssetRequest() bool {
	return resp.Type == assetRequest
}

func (resp ShopifyResponse) isListRequest() bool {
	return resp.Type == listRequest
}

func (resp ShopifyResponse) String() string {
	return fmt.Sprintf(`[%s] Performed %s at %s
	Request: %s
	Theme: %s
	Asset: %s
	Assets: %s
	Errors: %s`,
		RedText(resp.Code),
		YellowText(resp.EventType),
		YellowText(resp.Host),
		YellowText(resp.URL),
		YellowText(resp.Theme),
		YellowText(resp.Asset),
		YellowText(resp.Assets),
		resp.Errors,
	)
}

func (resp ShopifyResponse) Error() Error {
	if !resp.Successful() {
		if resp.isThemeRequest() {
			return themeError{resp}
		} else if resp.isAssetRequest() {
			return assetError{resp}
		} else if resp.isListRequest() {
			return listError{resp}
		}
		return kitError{resp.Errors}
	}
	return nil
}
