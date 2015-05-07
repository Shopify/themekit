package themekit

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"sync"
	"time"
)

const CreateThemeMaxRetries int = 3

type ThemeClient struct {
	config Configuration
	client *http.Client
	filter EventFilter
}

type Theme struct {
	Name        string `json:"name"`
	Source      string `json:"src,omitempty"`
	Role        string `json:"role,omitempty"`
	Id          int64  `json:"id,omitempty"`
	Previewable bool   `json:"previewable,omitempty"`
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
	return ThemeClient{
		config: config,
		client: newHttpClient(config),
		filter: NewEventFilterFromPatternsAndFiles(config.IgnoredFiles, config.Ignores),
	}
}

func (t ThemeClient) GetConfiguration() Configuration {
	return t.config
}

func (t ThemeClient) AssetList() (results chan Asset, errs chan error) {
	results = make(chan Asset)
	errs = make(chan error)
	go func() {
		queryBuilder := func(path string) string {
			return path
		}

		bytes, err := t.query(queryBuilder)
		if err != nil {
			errs <- err
			return
		}

		var assets map[string][]Asset
		err = json.Unmarshal(bytes, &assets)
		if err != nil {
			errs <- err
			return
		}

		for _, asset := range assets["assets"] {
			results <- asset
		}
		close(results)
		close(errs)
	}()
	return
}

type AssetRetrieval func(filename string) (Asset, error)

func (t ThemeClient) Asset(filename string) (Asset, error) {
	queryBuilder := func(path string) string {
		return fmt.Sprintf("%s&asset[key]=%s", path, filename)
	}

	bytes, err := t.query(queryBuilder)
	var asset map[string]Asset
	err = json.Unmarshal(bytes, &asset)
	if err != nil {
		return Asset{}, err
	}

	return asset["asset"], nil
}

func (t ThemeClient) CreateTheme(name, zipLocation string) (ThemeClient, chan ThemeEvent) {
	var wg sync.WaitGroup
	wg.Add(1)
	path := fmt.Sprintf("%s/themes.json", t.config.AdminUrl())
	contents := map[string]Theme{
		"theme": Theme{Name: name, Source: zipLocation, Role: "unpublished"},
	}

	log := make(chan ThemeEvent)
	logEvent := func(t ThemeEvent) {
		log <- t
	}

	retries := 0
	themeEvent := func() (themeEvent APIThemeEvent) {
		ready := false
		data, _ := json.Marshal(contents)
		for retries < CreateThemeMaxRetries && !ready {
			if themeEvent = t.sendData("POST", path, data); !themeEvent.Successful() {
				retries++
			} else {
				ready = true
			}
			go logEvent(themeEvent)
		}
		if retries >= CreateThemeMaxRetries {
			err := errors.New(fmt.Sprintf("'%s' cannot be retrieved from Github.", zipLocation))
			NotifyError(err)
		}
		return
	}()

	go func() {
		for !t.isDoneProcessing(themeEvent.ThemeId) {
			time.Sleep(250 * time.Millisecond)
		}
		wg.Done()
	}()

	wg.Wait()
	config := t.GetConfiguration()
	config.ThemeId = themeEvent.ThemeId
	return NewThemeClient(config.Initialize()), log
}

func (t ThemeClient) Process(events chan AssetEvent) (done chan bool, messages chan ThemeEvent) {
	done = make(chan bool)
	messages = make(chan ThemeEvent)
	go func() {
		for {
			job, more := <-events
			if more {
				if !t.filter.MatchesFilter(job.Asset().Key) {
					messages <- t.Perform(job)
				}
			} else {
				close(messages)
				done <- true
				return
			}
		}
	}()
	return
}

func (t ThemeClient) Perform(asset AssetEvent) ThemeEvent {
	var event string
	switch asset.Type() {
	case Update:
		event = "PUT"
	case Remove:
		event = "DELETE"
	}
	resp, err := t.request(asset, event)
	if err == nil {
		defer resp.Body.Close()
	}
	return processResponse(resp, err, asset)
}

func (t ThemeClient) query(queryBuilder func(path string) string) ([]byte, error) {
	path := fmt.Sprintf("%s?fields=key,attachment,value", t.config.AssetPath())
	path = queryBuilder(path)

	req, err := http.NewRequest("GET", path, nil)
	if err != nil {
		return []byte{}, err
	}

	t.config.AddHeaders(req)
	resp, err := t.client.Do(req)
	if err != nil {
		return []byte{}, err
	} else {
		defer resp.Body.Close()
	}
	return ioutil.ReadAll(resp.Body)
}

func (t ThemeClient) sendData(method, path string, body []byte) (result APIThemeEvent) {
	req, err := http.NewRequest(method, path, bytes.NewBuffer(body))
	if err != nil {
		NotifyError(err)
	}
	t.config.AddHeaders(req)
	resp, err := t.client.Do(req)
	if result = NewAPIThemeEvent(resp, err); result.Successful() {
		defer resp.Body.Close()
	}
	return result
}

func (t ThemeClient) request(event AssetEvent, method string) (*http.Response, error) {
	path := t.config.AssetPath()
	data := map[string]Asset{"asset": event.Asset()}
	encoded, err := json.Marshal(data)

	req, err := http.NewRequest(method, path, bytes.NewBuffer(encoded))

	if err != nil {
		return nil, err
	}

	t.config.AddHeaders(req)
	return t.client.Do(req)
}

func processResponse(r *http.Response, err error, event AssetEvent) ThemeEvent {
	return NewAPIAssetEvent(r, event, err)
}

func (t ThemeClient) isDoneProcessing(themeId int64) bool {
	path := fmt.Sprintf("%s/themes/%d.json", t.config.AdminUrl(), themeId)
	themeEvent := t.sendData("GET", path, []byte{})
	return themeEvent.Previewable
}

func ExtractErrorMessage(data []byte, err error) string {
	return extractAssetAPIErrors(data, err).Error()
}

func newHttpClient(config Configuration) (client *http.Client) {
	client = &http.Client{}
	if len(config.Proxy) > 0 {
		fmt.Println("Proxy URL detected from Configuration:", config.Proxy)
		fmt.Println("SSL Certificate Validation will be disabled!")
		proxyUrl, err := url.Parse(config.Proxy)
		if err != nil {
			fmt.Println("Proxy configuration invalid:", err)
		}
		client.Transport = &http.Transport{Proxy: http.ProxyURL(proxyUrl), TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}
	}
	return
}
