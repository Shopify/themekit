package kit

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"runtime"
)

type shopifyRequest struct {
	*http.Request
	config *Configuration
	url    string
	event  EventType
	rtype  requestType
	body   io.Reader
}

func newShopifyRequest(config *Configuration, rtype requestType, event EventType, urlStr string) *shopifyRequest {
	req := &shopifyRequest{
		config: config,
		url:    urlStr,
		event:  event,
		rtype:  rtype,
	}
	req.generateRequest()
	return req
}

func (req *shopifyRequest) setBody(body io.Reader) {
	req.body = body
	req.generateRequest()
}

func (req *shopifyRequest) setJSONBody(body map[string]interface{}) {
	data, _ := json.Marshal(body)
	req.setBody(bytes.NewBuffer(data))
}

func (req *shopifyRequest) generateRequest() {
	req.Request, _ = http.NewRequest(req.event.toMethod(), req.url, req.body)
	req.Header.Add("X-Shopify-Access-Token", req.config.Password)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")
	req.Header.Add("User-Agent", fmt.Sprintf("go/themekit (%s; %s; %s)", runtime.GOOS, runtime.GOARCH, ThemeKitVersion.String()))
}
