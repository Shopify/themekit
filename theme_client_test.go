package themekit

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"sort"
	"testing"

	"github.com/Shopify/themekit/theme"
	"github.com/stretchr/testify/assert"
)

type TestEvent struct {
	asset     theme.Asset
	eventType EventType
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

	client := NewThemeClient(config)
	client.Perform(event)
}

func TestProcessingAnEventsChannel(t *testing.T) {
	results := map[string]int{}
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		results[r.Method]++
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
	assets, _ := client.AssetList()
	assert.Equal(t, 2, count(assets))
}

func TestRetrievingLocalAssets(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	client := NewThemeClient(conf(ts))

	dir, _ := os.Getwd()
	assets := client.LocalAssets(fmt.Sprintf("%s/fixtures/templates", dir))

	assert.Equal(t, 1, len(assets))
}

func TestRetrievingLocalAssetsWithSubdirectories(t *testing.T) {
	client := NewThemeClient(Configuration{})

	dir, _ := os.Getwd()
	assets := client.LocalAssets(fmt.Sprintf("%s/fixtures/local_assets", dir))

	assert.Equal(t, 3, len(assets))
}

func TestRetrievingAnAssetListThatIncludesCompiledAssets(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, TestFixture("assets_response_from_shopify"))
	}))

	var expected map[string][]theme.Asset
	json.Unmarshal(RawTestFixture("expected_asset_list_output"), &expected)
	sort.Sort(theme.ByAsset(expected["assets"]))

	client := NewThemeClient(conf(ts))
	assetsChan, _ := client.AssetList()
	actual := makeSlice(assetsChan)
	sort.Sort(theme.ByAsset(actual))

	assert.Equal(t, len(expected["assets"]), len(actual))
	for index, expected := range expected["assets"] {
		assert.Equal(t, expected, actual[index])
	}
}

func TestRetrievingASingleAsset(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "fields=key,attachment,value&asset[key]=assets/foo.txt", r.URL.RawQuery)
		fmt.Fprintf(w, TestFixture("response_single"))
	}))

	client := NewThemeClient(conf(ts))
	asset, _ := client.Asset("assets/foo.txt")
	assert.Equal(t, "hello world", asset.Value)
}

func TestExtractErrorMessage(t *testing.T) {
	contents := []byte(TestFixture("asset_error"))
	expectedMessage := "Liquid syntax error (line 10): 'comment' tag was never closed"
	assert.Equal(t, expectedMessage, ExtractErrorMessage(contents, nil))
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

func TestThemeClientAssetListOnUnauthorized(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
	}))
	client := NewThemeClient(conf(ts))

	_, errs := client.AssetList()

	err := <-errs
	assert.NotNil(t, err)
	assert.Equal(t, "Server responded with HTTP 401; please check your credentials.", err.Error())
}

func asset() theme.Asset {
	return theme.Asset{Key: "assets/hello.txt", Value: "Hello World"}
}

func conf(server *httptest.Server) Configuration {
	return Configuration{Url: server.URL, AccessToken: "abra"}
}

func drain(channel chan ThemeEvent) {
	for {
		_, more := <-channel
		if !more {
			return
		}
	}
}

func count(channel chan theme.Asset) int {
	count := 0
	for {
		_, more := <-channel
		if !more {
			return count
		}
		count++
	}
}

func makeSlice(channel chan theme.Asset) []theme.Asset {
	assets := []theme.Asset{}
	for {
		asset, more := <-channel
		if !more {
			return assets
		}
		assets = append(assets, asset)
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
