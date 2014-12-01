package phoenix

import (
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
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
	ts := assertRequest(t, "PUT", "asset", map[string]string{"value": "Hello World", "key": "assets/hello.txt"})
	defer ts.Close()
	asset := TestEvent{asset: asset(), eventType: Update}
	client := NewThemeClient(conf(ts))
	client.Perform(asset)
}

func TestPerformWithRemoveAssetEvent(t *testing.T) {
	ts := assertRequest(t, "DELETE", "asset", map[string]string{"key": "assets/hello.txt"})
	defer ts.Close()
	asset := TestEvent{asset: asset(), eventType: Remove}
	client := NewThemeClient(conf(ts))
	client.Perform(asset)
}

func TestProcessingAnEventsChannel(t *testing.T) {
	results := map[string]int{}
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		results[r.Method] += 1
	}))

	stream := make(chan AssetEvent)
	go func() {
		stream <- TestEvent{asset: asset(), eventType: Update}
		stream <- TestEvent{asset: asset(), eventType: Update}
		stream <- TestEvent{asset: asset(), eventType: Remove}
		close(stream)
	}()

	client := NewThemeClient(conf(ts))
	done, messages := client.Process(stream)

	go drain(messages)

	<-done
	assert.Equal(t, 2, results["PUT"])
	assert.Equal(t, 1, results["DELETE"])
}

func TestRetrievingAnAssetList(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "fields=key,attachment,value", r.URL.RawQuery)
		fmt.Fprintf(w, TestFixture("response_multi"))
	}))

	client := NewThemeClient(conf(ts))
	assets := client.AssetList()
	assert.Equal(t, 2, count(assets))
}

func TestRetrievingASingleAsset(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "fields=key,attachment,value&asset[key]=assets/foo.txt", r.URL.RawQuery)
		fmt.Fprintf(w, TestFixture("response_single"))
	}))

	client := NewThemeClient(conf(ts))
	asset := client.Asset("assets/foo.txt")
	assert.Equal(t, "hello world", asset.Value)
}

func asset() Asset {
	return Asset{Key: "assets/hello.txt", Value: "Hello World"}
}

func conf(server *httptest.Server) Configuration {
	return Configuration{Url: server.URL, AccessToken: "abra"}
}

func drain(channel chan string) {
	for {
		_, more := <-channel
		if !more {
			return
		}
	}
}

func count(channel chan Asset) int {
	count := 0
	for {
		_, more := <-channel
		if !more {
			return count
		}
		count++
	}
}

func assertRequest(t *testing.T, method string, root string, formValues map[string]string) *httptest.Server {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "abra", r.Header.Get("X-Shopify-Access-Token"))
		assert.Equal(t, method, r.Method)
		var results map[string]map[string]string
		data, _ := ioutil.ReadAll(r.Body)
		json.Unmarshal(data, &results)
		values := results[root]
		for key, value := range formValues {
			assert.Equal(t, value, values[key])
		}
	}))
	return ts
}
