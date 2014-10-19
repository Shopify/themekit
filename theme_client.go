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
	Key        string `json:key`
	Value      string `json:value`
	Attachment string `json:attachment`
}

func (a Asset) String() string {
	return fmt.Sprintf("key: %s | value: %s | attachment: %s", a.Key, a.Value, a.Attachment)
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
	results := make(chan Asset)
	go func() {
		path := fmt.Sprintf("%s?fields=key,attachment,value", t.config.AssetPath())

		req, err := http.NewRequest("GET", path, nil)
		if err != nil {
			log.Fatal("Invalid Request", err)
		}

		t.config.AddHeaders(req)
		resp, err := t.client.Do(req)
		defer resp.Body.Close()
		if err != nil {
			log.Fatal("Invalid Response", err)
		}

		bytes, err := ioutil.ReadAll(resp.Body)
		var assets map[string][]Asset
		err = json.Unmarshal(bytes, &assets)
		if err != nil {
			log.Fatal(err)
		}

		for _, asset := range assets["assets"] {
			results <- asset
		}
		close(results)
	}()
	return results
}

type AssetRetrieval func(filename string) Asset

func (t ThemeClient) Asset(filename string) Asset {
	return Asset{}
}

func (t ThemeClient) Process(events chan AssetEvent) (done chan bool, messages chan string) {
	done = make(chan bool)
	messages = make(chan string)
	go func() {
		for {
			job, more := <-events
			if more {
				resp, err := t.Perform(job)
				messages <- processResponse(resp, err, job)
			} else {
				close(messages)
				done <- true
				return
			}
		}
	}()
	return
}

func (t ThemeClient) Perform(asset AssetEvent) (response *http.Response, err error) {
	var event string
	switch asset.Type() {
	case Update:
		event = "PUT"
	case Remove:
		event = "DELETE"
	}
	response, err = t.request(asset, event)
	if err != nil {
		log.Fatal(err)
	}
	return
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

func processResponse(r *http.Response, err error, event AssetEvent) string {
	asset := event.Asset()
	if err != nil {
		return err.Error()
	}
	host := BlueText(r.Request.URL.Host)
	key := BlueText(asset.Key)
	eventType := YellowText(event.Type().String())
	code := r.StatusCode
	if code >= 200 && code < 300 {
		return fmt.Sprintf("Successfully performed %s operation for file %s to %s", eventType, key, host)
	} else {
		return fmt.Sprintf("[%d]Could not peform %s to %s at %s", code, eventType, key, host)
	}
}
