package phoenix

import (
	"bytes"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

const CreateThemeMaxRetries int = 3

type ThemeClient struct {
	config Configuration
	client *http.Client
}

type Asset struct {
	Key        string `json:"key"`
	Value      string `json:"value,omitempty"`
	Attachment string `json:"attachment,omitempty"`
}

type Theme struct {
	Name        string `json:"name"`
	Source      string `json:"src,omitempty"`
	Role        string `json:"role,omitempty"`
	Id          int64  `json:"id,omitempty"`
	Previewable bool   `json:"previewable,omitempty"`
}

func (a Asset) String() string {
	return fmt.Sprintf("key: %s | value: %s | attachment: %s", a.Key, a.Value, a.Attachment)
}

func (a Asset) IsValid() bool {
	return len(a.Key) > 0 && (len(a.Value) > 0 || len(a.Attachment) > 0)
}

func toSlash(path string) string {
	newpath := filepath.ToSlash(path)
	if strings.Index(newpath, "\\") >= 0 {
		newpath = strings.Replace(newpath, "\\", "/", -1)
	}
	return newpath
}

func LoadAsset(root, filename string) (asset Asset, err error) {
	path := toSlash(fmt.Sprintf("%s/%s", root, filename))
	file, err := os.Open(path)
	info, err := os.Stat(path)
	if err != nil {
		return
	}

	if info.IsDir() {
		err = errors.New("File is a directory")
		return
	}

	buffer := make([]byte, info.Size())
	_, err = file.Read(buffer)
	if err != nil {
		return
	}

	asset = Asset{Key: toSlash(filename)}
	if contentTypeFor(buffer) == "text" {
		asset.Value = string(buffer)
	} else {
		asset.Attachment = encode64(buffer)
	}
	return
}

func contentTypeFor(data []byte) string {
	contentType := http.DetectContentType(data)
	if strings.Contains(contentType, "text") {
		return "text"
	} else {
		return "binary"
	}
}

func encode64(data []byte) string {
	return base64.StdEncoding.EncodeToString(data)
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
	return ThemeClient{config: config, client: newHttpClient(config)}
}

func (t ThemeClient) GetConfiguration() Configuration {
	return t.config
}

func (t ThemeClient) AssetList() chan Asset {
	results := make(chan Asset)
	go func() {
		queryBuilder := func(path string) string {
			return path
		}

		bytes, err := t.query(queryBuilder)
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
	queryBuilder := func(path string) string {
		return fmt.Sprintf("%s&asset[key]=%s", path, filename)
	}

	bytes, err := t.query(queryBuilder)
	var asset map[string]Asset
	err = json.Unmarshal(bytes, &asset)
	if err != nil {
		log.Fatal(err)
	}

	return asset["asset"]
}

func (t ThemeClient) CreateTheme(name, zipLocation string) (tc ThemeClient) {
	var wg sync.WaitGroup
	wg.Add(1)
	path := fmt.Sprintf("%s/themes.json", t.config.AdminUrl())
	contents := map[string]Theme{
		"theme": Theme{Name: name, Source: zipLocation, Role: "unpublished"},
	}

	retries := 0
	themeEvent := func() (themeEvent APIThemeEvent) {
		ready := false
		data, _ := json.Marshal(contents)
		for retries < CreateThemeMaxRetries && !ready {
			if themeEvent = t.sendData("POST", path, data); !themeEvent.Successful() {
				retries++
				msg := fmt.Sprintf("[%d/%d] %s", retries, CreateThemeMaxRetries, themeEvent)
				fmt.Println(YellowText(msg))
			} else {
				ready = true
			}
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
	return NewThemeClient(config.Initialize())
}

func (t ThemeClient) Process(events chan AssetEvent) (done chan bool, messages chan ThemeEvent) {
	done = make(chan bool)
	messages = make(chan ThemeEvent)
	go func() {
		for {
			job, more := <-events
			if more {
				messages <- t.Perform(job)
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
		log.Fatal("Invalid Request", err)
	}

	t.config.AddHeaders(req)
	resp, err := t.client.Do(req)
	defer resp.Body.Close()
	if err != nil {
		log.Fatal("Invalid response", err)
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
		log.Fatal(err)
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
