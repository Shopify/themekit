package kit

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestShopfiyRequest(t *testing.T) {
	req := newShopifyRequest(&Configuration{Password: "sharknado"}, assetRequest, Update, "http://localhost:5000")
	assert.Equal(t, "sharknado", req.Header.Get("X-Shopify-Access-Token"))
	assert.Equal(t, "application/json", req.Header.Get("Content-Type"))
	assert.Equal(t, "application/json", req.Header.Get("Accept"))
	assert.Equal(t, fmt.Sprintf("go/themekit (%s; %s; %s)", runtime.GOOS, runtime.GOARCH, ThemeKitVersion), req.Header.Get("User-Agent"))

	reader := bytes.NewBufferString("my string")
	req.setBody(reader)
	assert.Equal(t, reader, req.body)

	req.setJSONBody(map[string]interface{}{"assest": Asset{Key: "hello.txt"}})
	out, _ := ioutil.ReadAll(req.body)
	assert.Equal(t, `{"assest":{"key":"hello.txt"}}`, string(out))
}
