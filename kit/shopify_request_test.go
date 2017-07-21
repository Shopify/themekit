package kit

import (
	"bytes"
	"fmt"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewRequest(t *testing.T) {
	req, err := newShopifyRequest(&Configuration{Password: "sharknado"}, assetRequest, Update, "http://localhost:5000")
	assert.Nil(t, err)
	assert.Equal(t, "sharknado", req.Header.Get("X-Shopify-Access-Token"))
	assert.Equal(t, "application/json", req.Header.Get("Content-Type"))
	assert.Equal(t, "application/json", req.Header.Get("Accept"))
	assert.Equal(t, fmt.Sprintf("go/themekit (%s; %s; %s)", runtime.GOOS, runtime.GOARCH, ThemeKitVersion), req.Header.Get("User-Agent"))

	_, err = newShopifyRequest(&Configuration{}, assetRequest, Update, "://#nksd")
	assert.NotNil(t, err)
}

func TestSetBody(t *testing.T) {
	req, _ := newShopifyRequest(&Configuration{}, assetRequest, Update, "http://localhost:5000")
	reader := bytes.NewBufferString("my string")
	err := req.setBody(reader)
	assert.Nil(t, err)
	assert.Equal(t, reader, req.body)
}

func TestSetJSONBody(t *testing.T) {
	req, _ := newShopifyRequest(&Configuration{}, assetRequest, Update, "http://localhost:5000")
	err := req.setJSONBody(map[string]interface{}{"assest": Asset{Key: "hello.txt"}})
	assert.Nil(t, err)
}
