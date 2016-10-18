package kit

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/Shopify/themekit/theme"
)

type ShopifyResponse struct {
	Type      requestType   `json:"-"`
	Host      string        `json:"host"`
	URL       *url.URL      `json:"url"`
	Code      int           `json:"status_code"`
	Theme     theme.Theme   `json:"theme"`
	Asset     theme.Asset   `json:"asset"`
	Assets    []theme.Asset `json:"assets"`
	EventType EventType     `json:"event_type"`
	Errors    string        `json:"errors"`
}

func newShopifyResponse(rtype requestType, event EventType, resp *http.Response, err error) (*ShopifyResponse, Error) {
	if resp == nil || err != nil {
		return nil, KitError{err}
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
		return nil, KitError{err}
	}

	json.Unmarshal(bytes, &newResponse)

	return newResponse, newResponse.Error()
}

func (resp ShopifyResponse) Successful() bool {
	return resp.Code >= 200 && resp.Code < 300 && len(resp.Errors) == 0
}

func (resp ShopifyResponse) IsThemeRequest() bool {
	return resp.Type == themeRequest
}

func (resp ShopifyResponse) IsAssetRequest() bool {
	return resp.Type == assetRequest
}

func (resp ShopifyResponse) IsListRequest() bool {
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
		if resp.IsThemeRequest() {
			return ThemeError{resp}
		} else if resp.IsAssetRequest() {
			return AssetError{resp}
		} else if resp.IsListRequest() {
			return ListError{resp}
		} else {
			return KitError{fmt.Errorf(resp.Errors)}
		}
	}
	return nil
}
