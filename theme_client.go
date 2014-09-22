package phoenix

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
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

func (e EventType) String() string {
	switch e {
	case Update:
		return "Update"
	case Remove:
		return "Remove"
	default:
		return "Unknown"
	}
}

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
	var event string
	switch asset.Type() {
	case Update:
		event = "PUT"
	case Remove:
		event = "DELETE"
	}
	_, err := t.request(asset, event)
	if err != nil {
		log.Fatal(err)
	}
}

func (t ThemeClient) request(event AssetEvent, method string) (*http.Response, error) {
	path := t.config.AssetPath()
	data := map[string]map[string]string{"asset": {"key": event.Asset().Key, "value": event.Asset().Value}}
	encoded, err := json.Marshal(data)

	req, err := http.NewRequest(method, path, strings.NewReader(string(encoded)))

	if err != nil {
		log.Fatal(err)
	}

	t.config.AddHeaders(req)
	return t.client.Do(req)
}
