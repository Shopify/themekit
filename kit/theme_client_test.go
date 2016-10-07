package kit

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"image/png"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"sort"
	"testing"

	"github.com/Shopify/themekit/theme"
	"github.com/stretchr/testify/assert"
)

var globalEventLog = make(chan ThemeEvent, 100)

type TestEvent struct {
	asset     theme.Asset
	eventType EventType
}

func Fixture(name string) string {
	return string(RawFixture(name))
}

func RawFixture(name string) []byte {
	path := fmt.Sprintf("../fixtures/%s.json", name)
	file, err := os.Open(path)
	defer file.Close()
	if err != nil {
		log.Fatal(err)
	}
	bytes, err := ioutil.ReadAll(file)
	if err != nil {
		log.Fatal(err)
	}
	return bytes
}

func BinaryTestData() []byte {
	img := image.NewRGBA(image.Rect(0, 0, 10, 10))
	buff := bytes.NewBuffer([]byte{})
	png.Encode(buff, img)
	return buff.Bytes()
}

func (t TestEvent) Asset() theme.Asset {
	return t.asset
}

func (t TestEvent) Type() EventType {
	return t.eventType
}

func TestPerformWithUpdateAssetEvent(t *testing.T) {
	ts := assertRequest(t, "PUT", "asset", map[string]string{"value": "Hello World", "key": "assets/hello.txt"})
	defer ts.Close()
	asset := TestEvent{asset: asset(), eventType: Update}
	client := NewThemeClient(globalEventLog, conf(ts))
	client.Perform(asset)
}

func TestPerformWithRemoveAssetEvent(t *testing.T) {
	ts := assertRequest(t, "DELETE", "asset", map[string]string{"key": "assets/hello.txt"})
	defer ts.Close()
	asset := TestEvent{asset: asset(), eventType: Remove}
	client := NewThemeClient(globalEventLog, conf(ts))
	client.Perform(asset)
}

func TestPerformWithAssetEventThatDoesNotPassTheFilter(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Log("The request should never have been sent")
		t.Fail()
	}))
	defer ts.Close()
	config := conf(ts)
	config.IgnoredFiles = []string{"snickerdoodle.txt"}
	config.Ignores = []string{}

	asset := theme.Asset{Key: "snickerdoodle.txt", Value: "not important"}
	event := TestEvent{asset: asset, eventType: Update}

	client := NewThemeClient(globalEventLog, config)
	client.Perform(event)
}

func TestProcessingAnEventsChannel(t *testing.T) {
	results := map[string]int{}
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		results[r.Method]++
	}))

	client := NewThemeClient(globalEventLog, conf(ts))

	done := make(chan bool)
	stream := client.Process(done)
	go func() {
		stream <- TestEvent{asset: asset(), eventType: Update}
		stream <- TestEvent{asset: asset(), eventType: Update}
		stream <- TestEvent{asset: asset(), eventType: Remove}
		close(stream)
	}()

	<-done
	assert.Equal(t, 2, results["PUT"])
	assert.Equal(t, 1, results["DELETE"])
}

func TestRetrievingAnAssetList(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "fields=key,attachment,value", r.URL.RawQuery)
		fmt.Fprintf(w, Fixture("response_multi"))
	}))

	client := NewThemeClient(globalEventLog, conf(ts))
	assets := client.AssetList()
	assert.Equal(t, 2, len(assets))
}

func TestRetrievingLocalAssets(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	client := NewThemeClient(globalEventLog, conf(ts))

	assets := client.LocalAssets("../fixtures/templates")

	assert.Equal(t, 1, len(assets))
}

func TestRetrievingLocalAssetsWithSubdirectories(t *testing.T) {
	client := NewThemeClient(globalEventLog, Configuration{})

	assets := client.LocalAssets("../fixtures/local_assets")

	assert.Equal(t, 3, len(assets))
}

func TestRetrievingAnAssetListThatIncludesCompiledAssets(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, Fixture("assets_response_from_shopify"))
	}))

	var expected map[string][]theme.Asset
	json.Unmarshal(RawFixture("expected_asset_list_output"), &expected)
	sort.Sort(theme.ByAsset(expected["assets"]))

	client := NewThemeClient(globalEventLog, conf(ts))
	assetsChan := client.AssetList()
	sort.Sort(theme.ByAsset(assetsChan))

	assert.Equal(t, len(expected["assets"]), len(assetsChan))
	for index, expected := range expected["assets"] {
		assert.Equal(t, expected, assetsChan[index])
	}
}

func TestRetrievingASingleAsset(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "fields=key,attachment,value&asset[key]=assets/foo.txt", r.URL.RawQuery)
		fmt.Fprintf(w, Fixture("response_single"))
	}))

	client := NewThemeClient(globalEventLog, conf(ts))
	asset, _ := client.Asset("assets/foo.txt")
	assert.Equal(t, "hello world", asset.Value)
}

func TestIgnoringCompiledAssets(t *testing.T) {
	input := []theme.Asset{
		{Key: "assets/ajaxify.js"},
		{Key: "assets/ajaxify.js.liquid"},
		{Key: "assets/checkout.css"},
		{Key: "assets/checkout.css.liquid"},
		{Key: "templates/article.liquid"},
		{Key: "templates/product.liquid"},
	}
	expected := []theme.Asset{
		{Key: "assets/ajaxify.js.liquid"},
		{Key: "assets/checkout.css.liquid"},
		{Key: "templates/article.liquid"},
		{Key: "templates/product.liquid"},
	}
	assert.Equal(t, expected, ignoreCompiledAssets(input))
}

func asset() theme.Asset {
	return theme.Asset{Key: "assets/hello.txt", Value: "Hello World"}
}

func conf(server *httptest.Server) Configuration {
	return Configuration{
		URL:         server.URL,
		AccessToken: "abra",
		BucketSize:  100,
		RefillRate:  100,
	}
}

func drain(channel chan ThemeEvent) {
	for {
		_, more := <-channel
		if !more {
			return
		}
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
