package phoenix

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
)

type ThemeClient struct {
	config Configuration
	client *http.Client
}

type Asset struct {
	Key   string `json:key`
	Value string `json:value`
}

type EventType int

const (
	Update EventType = iota
	Remove
)

type AssetEvent interface {
	Asset() Asset
	Type() EventType
}

func NewThemeClient(config Configuration) ThemeClient {
	return ThemeClient{config: config, client: &http.Client{}}
}

func (t ThemeClient) AssetList() chan Asset {
	assets := make(chan Asset)
	go func() {
		path := fmt.Sprintf("%s?fields=key,attachment", t.config.AssetPath())
		req, err := http.NewRequest("GET", path, nil)
		resp, err := t.client.Do(req)
		defer resp.Body.Close()
		bytes, err := ioutil.ReadAll(resp.Body)
		var assets map[string]Asset
		err = json.Unmarshal(bytes, &assets)
		if err == nil {

		}
	}()
	return assets
}

func (t ThemeClient) Process(events chan AssetEvent) {
	go func() {
		for {
			job, more := <-events
			if more {
				t.Perform(job)
			} else {
				return
			}
		}
	}()
}

func (t ThemeClient) Perform(asset AssetEvent) {
	switch asset.Type() {
	case Update:
		t.request(asset, "PUT")
	case Remove:
		t.request(asset, "DELETE")
	}
}

func (t ThemeClient) request(event AssetEvent, method string) (*http.Response, error) {
	path := t.config.AssetPath()
	content := url.Values{}
	content.Set("asset[key]", event.Asset().Key)
	content.Set("asset[value]", event.Asset().Value)

	req, err := http.NewRequest(method, path, strings.NewReader(content.Encode()))

	if err != nil {
		log.Fatal(err)
	}

	t.config.AddHeaders(req)
	return t.client.Do(req)
}
