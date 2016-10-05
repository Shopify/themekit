package kit

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/Shopify/themekit/theme"
)

const createThemeMaxRetries int = 3

// ThemeClient ... TODO
type ThemeClient struct {
	eventLog   chan ThemeEvent
	config     Configuration
	httpClient *http.Client
	filter     eventFilter
}

type apiResponse struct {
	code int
	body []byte
	err  error
}

// EventType ... TODO
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

// NonFatalNetworkError ... TODO
type NonFatalNetworkError struct {
	Code    int
	Verb    string
	Message string
}

func (e NonFatalNetworkError) Error() string {
	return fmt.Sprintf("%d %s %s", e.Code, e.Verb, e.Message)
}

const (
	// Update ... TODO
	Update EventType = iota
	// Remove ... TODO
	Remove
)

// AssetEvent ... TODO
type AssetEvent interface {
	Asset() theme.Asset
	Type() EventType
}

// NewThemeClient ... TODO
func NewThemeClient(eventLog chan ThemeEvent, config Configuration) ThemeClient {
	return ThemeClient{
		eventLog:   eventLog,
		config:     config,
		httpClient: newHTTPClient(config),
		filter:     newEventFilterFromPatternsAndFiles(config.IgnoredFiles, config.Ignores),
	}
}

// GetConfiguration ... TODO
func (t ThemeClient) GetConfiguration() Configuration {
	return t.config
}

// NewFileWatcher creates a new filewatcher using the theme clients file filter
func (t ThemeClient) NewFileWatcher(dir, notifyFile string) chan AssetEvent {
	new_foreman := newForeman(newLeakyBucket(t.config.BucketSize, t.config.RefillRate, 1))
	if len(notifyFile) > 0 {
		new_foreman.OnIdle = func() {
			os.Create(notifyFile)
			os.Chtimes(notifyFile, time.Now(), time.Now())
		}
	}
	var err error
	new_foreman.JobQueue, err = newFileWatcher(dir, true, t.filter)
	if err != nil {
		Fatal(err)
	}
	new_foreman.Restart()
	return new_foreman.WorkerQueue
}

func (t ThemeClient) ErrorMessage(content string, args ...interface{}) {
	go func() {
		t.eventLog <- basicEvent{
			Formatter: func(b basicEvent) string { return RedText(fmt.Sprintf(content, args...)) },
			EventType: "message",
			Title:     "Notice",
			Etype:     "basicEvent",
		}
	}()
}

func (t ThemeClient) Message(content string, args ...interface{}) {
	go func() {
		t.eventLog <- basicEvent{
			Formatter: func(b basicEvent) string { return fmt.Sprintf(content, args...) },
			EventType: "message",
			Title:     "Notice",
			Etype:     "basicEvent",
		}
	}()
}

// AssetList ... TODO
func (t ThemeClient) AssetList() []theme.Asset {
	queryBuilder := func(path string) string {
		return path
	}

	resp := t.query(queryBuilder)
	if resp.err != nil {
		t.ErrorMessage(resp.err.Error())
	}

	if resp.code >= 400 && resp.code < 500 {
		t.ErrorMessage("Server responded with HTTP %d; please check your credentials.", resp.code)
		return []theme.Asset{}
	}
	if resp.code >= 500 {
		t.ErrorMessage("Server responded with HTTP %d; try again in a few minutes.", resp.code)
		return []theme.Asset{}
	}

	var assets map[string][]theme.Asset
	err := json.Unmarshal(resp.body, &assets)
	if err != nil {
		t.ErrorMessage(err.Error())
		return []theme.Asset{}
	}

	sort.Sort(theme.ByAsset(assets["assets"]))

	return t.filter.FilterAssets(ignoreCompiledAssets(assets["assets"]))
}

// LocalAssets ... TODO
func (t ThemeClient) LocalAssets(dir string) []theme.Asset {
	dir = fmt.Sprintf("%s%s", dir, string(filepath.Separator))

	assets, err := theme.LoadAssetsFromDirectory(dir, t.filter.MatchesFilter)
	if err != nil {
		panic(err)
	}

	return assets
}

// Asset ... TODO
func (t ThemeClient) Asset(filename string) (theme.Asset, error) {
	queryBuilder := func(path string) string {
		return fmt.Sprintf("%s&asset[key]=%s", path, filename)
	}

	resp := t.query(queryBuilder)
	if resp.err != nil {
		return theme.Asset{}, resp.err
	}
	if resp.code >= 400 {
		return theme.Asset{}, NonFatalNetworkError{Code: resp.code, Verb: "GET", Message: "not found"}
	}
	var asset map[string]theme.Asset
	err := json.Unmarshal(resp.body, &asset)
	if err != nil {
		return theme.Asset{}, err
	}

	return asset["asset"], nil
}

// CreateTheme ... TODO
func (t ThemeClient) CreateTheme(name, zipLocation string) ThemeClient {
	var wg sync.WaitGroup
	wg.Add(1)
	path := fmt.Sprintf("%s/themes.json", t.config.AdminURL())
	contents := map[string]theme.Theme{
		"theme": {Name: name, Source: zipLocation, Role: "unpublished"},
	}

	retries := 0
	themeEvent := func() (themeEvent APIThemeEvent) {
		ready := false
		data, _ := json.Marshal(contents)
		for retries < createThemeMaxRetries && !ready {
			if themeEvent = t.sendData("POST", path, data); !themeEvent.Successful() {
				retries++
			} else {
				ready = true
			}
			go func(client ThemeClient, event ThemeEvent) {
				client.eventLog <- event
			}(t, themeEvent)
		}
		if retries >= createThemeMaxRetries {
			err := fmt.Errorf(fmt.Sprintf("'%s' cannot be retrieved from Github.", zipLocation))
			Fatal(err)
		}
		return
	}()

	go func() {
		for !t.isDoneProcessing(themeEvent.ThemeID) {
			time.Sleep(250 * time.Millisecond)
		}
		wg.Done()
	}()

	wg.Wait()
	config := t.GetConfiguration() // Shouldn't this configuration already be loaded and initialized?
	config.ThemeID = fmt.Sprintf("%d", themeEvent.ThemeID)
	config, err := config.Initialize()
	if err != nil {
		// TODO: there's no way we can signal that something went wrong.
	}
	return NewThemeClient(t.eventLog, config)
}

// ProcessSync ... TODO
func (t ThemeClient) ProcessSync(events []AssetEvent) {
	for _, event := range events {
		t.Perform(event)
	}
}

// Process ... TODO
func (t ThemeClient) Process(done chan bool) chan AssetEvent {
	new_foreman := newForeman(newLeakyBucket(t.config.BucketSize, t.config.RefillRate, 1))
	go func() {
		for {
			job, more := <-new_foreman.WorkerQueue
			if more {
				t.Perform(job)
			} else {
				done <- true
				return
			}
		}
	}()
	return new_foreman.JobQueue
}

// Perform ... TODO
func (t ThemeClient) Perform(asset AssetEvent) {
	if t.filter.MatchesFilter(asset.Asset().Key) {
		return
	}

	var event string
	switch asset.Type() {
	case Update:
		event = "PUT"
	case Remove:
		event = "DELETE"
	}

	resp, err := t.request(asset, event)
	t.eventLog <- NewAPIAssetEvent(resp, asset, err)
}

func (t ThemeClient) query(queryBuilder func(path string) string) apiResponse {
	path := fmt.Sprintf("%s?fields=key,attachment,value", t.config.AssetPath())
	path = queryBuilder(path)

	req, err := http.NewRequest("GET", path, nil)
	if err != nil {
		return apiResponse{err: err}
	}

	t.config.AddHeaders(req)
	resp, err := t.httpClient.Do(req)
	if err != nil {
		return apiResponse{err: err}
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	return apiResponse{code: resp.StatusCode, body: body, err: err}
}

func (t ThemeClient) sendData(method, path string, body []byte) (result APIThemeEvent) {
	req, err := http.NewRequest(method, path, bytes.NewBuffer(body))
	if err != nil {
		Fatal(err)
	}
	t.config.AddHeaders(req)
	resp, err := t.httpClient.Do(req)
	if result = NewAPIThemeEvent(resp, err); result.Successful() {
		defer resp.Body.Close()
	}
	return result
}

func (t ThemeClient) request(event AssetEvent, method string) (*http.Response, error) {
	path := t.config.AssetPath()
	data := map[string]theme.Asset{"asset": event.Asset()}

	encoded, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(method, path, bytes.NewBuffer(encoded))

	if err != nil {
		return nil, err
	}

	t.config.AddHeaders(req)
	return t.httpClient.Do(req)
}

func (t ThemeClient) isDoneProcessing(themeID int64) bool {
	path := fmt.Sprintf("%s/themes/%d.json", t.config.AdminURL(), themeID)
	themeEvent := t.sendData("GET", path, []byte{})
	return themeEvent.Previewable
}

func newHTTPClient(config Configuration) (client *http.Client) {
	client = &http.Client{}
	if len(config.Proxy) > 0 {
		fmt.Println("Proxy URL detected from Configuration:", config.Proxy)
		fmt.Println("SSL Certificate Validation will be disabled!")
		proxyURL, err := url.Parse(config.Proxy)
		if err != nil {
			fmt.Println("Proxy configuration invalid:", err)
		}
		client.Transport = &http.Transport{Proxy: http.ProxyURL(proxyURL), TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}
	}
	return
}

func ignoreCompiledAssets(assets []theme.Asset) []theme.Asset {
	newSize := 0
	results := make([]theme.Asset, len(assets))
	isCompiled := func(a theme.Asset, rest []theme.Asset) bool {
		for _, other := range rest {
			if strings.Contains(other.Key, a.Key) {
				return true
			}
		}
		return false
	}
	for index, asset := range assets {
		if !isCompiled(asset, assets[index+1:]) {
			results[newSize] = asset
			newSize++
		}
	}
	return results[:newSize]
}
