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

func newShopifyRequest(config *Configuration, rtype requestType, event EventType, urlStr string) (*shopifyRequest, error) {
	req := &shopifyRequest{
		config: config,
		url:    urlStr,
		event:  event,
		rtype:  rtype,
	}
	return req, req.generateRequest()
}

func (req *shopifyRequest) setBody(body io.Reader) error {
	req.body = body
	return req.generateRequest()
}

func (req *shopifyRequest) setJSONBody(body map[string]interface{}) error {
	data, err := json.Marshal(body)
	if err != nil {
		return err
	}
	return req.setBody(bytes.NewBuffer(data))
}

func (req *shopifyRequest) generateRequest() error {
	var err error
	req.Request, err = http.NewRequest(req.event.toMethod(), req.url, req.body)
	if err != nil {
		return err
	}
	req.Header.Add("X-Shopify-Access-Token", req.config.Password)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")
	req.Header.Add("User-Agent", fmt.Sprintf("go/themekit (%s; %s; %s)", runtime.GOOS, runtime.GOARCH, ThemeKitVersion.String()))
	return nil
}
