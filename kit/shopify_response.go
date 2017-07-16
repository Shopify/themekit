package kit

import (
	"encoding/json"
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

func newShopifyResponse(req *shopifyRequest, resp *http.Response, err error) (*ShopifyResponse, Error) {
	parsedURL, _ := url.Parse(req.url)

	newResponse := &ShopifyResponse{
		Type:      req.rtype,
		EventType: req.event,
		Host:      parsedURL.Host,
		URL:       parsedURL,
	}
	newResponse.Errors.AddE(err)

	if resp != nil {
		defer resp.Body.Close()
		newResponse.Code = resp.StatusCode

		bytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			newResponse.Errors.AddE(err)
		} else {
			err = json.Unmarshal(bytes, &newResponse)
			if err != nil {
				reqErr := generalRequestError{}
				json.Unmarshal(bytes, &reqErr)
				newResponse.Errors.Add(reqErr)
			}
		}
	}

	return newResponse, newResponse.Error()
}

// Successful will return true if the response code >= 200 and < 300 and if no
// errors were returned from the server.
func (resp ShopifyResponse) Successful() bool {
	return resp.Code >= 200 && resp.Code < 300 && !resp.Errors.Any()
}

func (resp ShopifyResponse) Error() Error {
	if !resp.Successful() {
		switch resp.Type {
		case themeRequest:
			return newThemeError(resp)
		case assetRequest:
			return newAssetError(resp)
		case listRequest:
			return newListError(resp)
		default:
			return kitError{resp.Errors}
		}
	}
	return nil
}
