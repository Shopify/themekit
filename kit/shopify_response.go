package kit

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/Shopify/themekit/theme"
)

type shopifyResponse struct {
	Host      string              `json:"host"`
	Code      int                 `json:"status_code"`
	Theme     theme.Theme         `json:"theme"`
	Asset     theme.Asset         `json:"asset"`
	Assets    []theme.Asset       `json:"assets"`
	EventType EventType           `json:"event_type"`
	Errors    map[string][]string `json:"errors"`
}

func newShopifyResponse(event EventType, resp *http.Response, err error) (*shopifyResponse, error) {
	if resp == nil || err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	newResponse := &shopifyResponse{
		Host:      resp.Request.URL.Host,
		Code:      resp.StatusCode,
		EventType: event,
	}
	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(bytes, &newResponse)
	return newResponse, err
}

func (resp shopifyResponse) Successful() bool {
	return resp.Code >= 200 && resp.Code < 300
}

func (resp shopifyResponse) IsThemeRequest() bool {
	return resp.Theme.Name != ""
}

func (resp shopifyResponse) IsAssetRequest() bool {
	return resp.Asset.Key != ""
}

func (resp shopifyResponse) IsListRequest() bool {
	return resp.Assets != nil
}

func (resp shopifyResponse) String() string {
	if resp.IsThemeRequest() {
		return resp.ThemeString()
	} else if resp.IsAssetRequest() {
		return resp.AssetString()
	} else if resp.IsListRequest() {
		return resp.ListString()
	}
	return fmt.Sprintf(
		"[%s] Performed %s to %s at %s\n\t%s",
		RedText(fmt.Sprintf("%d", resp.Code)),
		YellowText(resp.EventType),
		YellowText(resp.Host),
		resp.fmtErrors(),
	)
}

func (resp shopifyResponse) ThemeString() string {
	if resp.Successful() {
		return fmt.Sprintf(
			"[%s]Modifications made to theme '%s' with id of %s on shop %s",
			GreenText(fmt.Sprintf("%d", resp.Code)),
			BlueText(resp.Theme.Name),
			BlueText(fmt.Sprintf("%d", resp.Theme.ID)),
			YellowText(resp.Host),
		)
	}

	return fmt.Sprintf(
		"[%s]Encoutered error with request to %s\n\t%s",
		RedText(fmt.Sprintf("%d", resp.Code)),
		YellowText(resp.Host),
		resp.fmtErrors(),
	)
}

func (resp shopifyResponse) AssetString() string {
	if resp.Successful() {
		return fmt.Sprintf(
			"Successfully performed %s operation for file %s to %s",
			GreenText(resp.EventType),
			BlueText(resp.Asset.Key),
			YellowText(resp.Host),
		)
	} else if resp.Code == 422 {
		return RedText(fmt.Sprintf(
			"Could not upload %s:\n\t%s",
			resp.Asset.Key,
			resp.fmtErrors(),
		))
	} else if resp.Code == 403 || resp.Code == 401 {
		return fmt.Sprintf(
			"[%s]Insufficient permissions to perform %s to %s",
			RedText(fmt.Sprintf("%d", resp.Code)),
			YellowText(resp.EventType),
			BlueText(resp.Asset.Key),
		)
	} else if resp.Code == 404 {
		return fmt.Sprintf(
			"[%s]Could not complete operation because %s does not exist",
			RedText(fmt.Sprintf("%d", resp.Code)),
			BlueText(resp.Asset.Key),
		)
	} else {
		return fmt.Sprintf(
			"[%s]Could not perform %s to %s at %s\n\t%s",
			RedText(fmt.Sprintf("%d", resp.Code)),
			YellowText(resp.EventType),
			BlueText(resp.Asset.Key),
			YellowText(resp.Host),
			resp.fmtErrors(),
		)
	}
	return ""
}

func (resp shopifyResponse) ListString() string {
	if resp.Code >= 400 && resp.Code < 500 {
		return fmt.Sprintf(
			"Server responded with HTTP %d; please check your credentials.",
			resp.Code,
		)
	} else if resp.Code >= 500 {
		return fmt.Sprintf(
			"Server responded with HTTP %d; try again in a few minutes.",
			resp.Code,
		)
	}
	return ""
}

func (resp shopifyResponse) fmtErrors() string {
	output := []string{}
	for attr, errors := range resp.Errors {
		output = append(output, fmt.Sprintf("%s error: %s", attr, strings.Join(errors, ",")))
	}
	return strings.Join(output, ",")
}
