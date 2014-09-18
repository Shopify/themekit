package phoenix

import (
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

type TestEvent struct {
	asset     Asset
	eventType EventType
}

func (t TestEvent) Asset() Asset {
	return t.asset
}

func (t TestEvent) Type() EventType {
	return t.eventType
}

func TestPerformWithUpdateAssetEvent(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "abra", r.Header.Get("X-Shopify-Access-Token"))
		assert.Equal(t, "PUT", r.Method)
		data, _ := ioutil.ReadAll(r.Body)
		v, _ := url.ParseQuery(string(data))
		assert.Equal(t, "Hello World", v.Get("asset[value]"))
		assert.Equal(t, "assets/hello.txt", v.Get("asset[key]"))
	}))
	defer ts.Close()
	asset := TestEvent{asset: asset(), eventType: Update}
	client := NewThemeClient(conf(ts))
	client.Perform(asset)
}

func asset() Asset {
	return Asset{Key: "assets/hello.txt", Value: "Hello World"}
}

func conf(server *httptest.Server) Configuration {
	return Configuration{Url: server.URL, AccessToken: "abra"}
}
