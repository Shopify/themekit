package httpify

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/Shopify/themekit/src/release"
	"github.com/stretchr/testify/assert"
)

func TestNewClient(t *testing.T) {
	_, err := NewClient(Params{
		Domain: "#$%$^@$%!#@$@bad url this will not parse",
	})
	if assert.NotNil(t, err) {
		assert.EqualError(t, err, "invalid domain #$%$^@$%!#@$@bad url this will not parse")
	}

	_, err = NewClient(Params{
		Domain: "http://localhost.com",
		Proxy:  "@#$@$^#!@#$@",
	})
	if assert.NotNil(t, err) {
		assert.EqualError(t, err, "invalid proxy URI")
	}
}

func TestClient_do(t *testing.T) {
	body := map[string]interface{}{"key": "main.js", "value": "alert('this is javascript');"}

	var client *HTTPClient
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, r.Header.Get("X-Custom-Header"), "Checksum")
		assert.Equal(t, r.Header.Get("X-Shopify-Access-Token"), client.password)
		assert.Equal(t, r.Header.Get("Content-Type"), "application/json")
		assert.Equal(t, r.Header.Get("Accept"), "application/json")
		assert.Equal(t, r.Header.Get("User-Agent"), fmt.Sprintf("go/themekit (%s; %s; %s)", runtime.GOOS, runtime.GOARCH, release.ThemeKitVersion.String()))

		reqBody, err := ioutil.ReadAll(r.Body)
		assert.Nil(t, err)

		asset := struct {
			Key   string `json:"key"`
			Value string `json:"value"`
		}{}
		assert.Nil(t, json.Unmarshal(reqBody, &asset))
		assert.Equal(t, asset.Key, body["key"])
		assert.Equal(t, asset.Value, body["value"])
	}))

	client, err := NewClient(Params{
		Domain:   server.URL,
		Password: "secret_password",
	})
	client.baseURL.Scheme = "http"

	assert.NotNil(t, client)
	assert.Nil(t, err)

	resp, err := client.Post("/assets.json", body, map[string]string{"X-Custom-Header": "Checksum"})
	assert.Nil(t, err)
	assert.NotNil(t, resp)

	resp, err = client.Put("/assets.json", body, map[string]string{"X-Custom-Header": "Checksum"})
	assert.Nil(t, err)
	assert.NotNil(t, resp)

	server.Close()

	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, r.Header.Get("X-Custom-Header"), "Foo")
		assert.Equal(t, r.Header.Get("X-Shopify-Access-Token"), client.password)
		assert.Equal(t, r.Header.Get("Content-Type"), "application/json")
		assert.Equal(t, r.Header.Get("Accept"), "application/json")
		assert.Equal(t, r.Header.Get("User-Agent"), fmt.Sprintf("go/themekit (%s; %s; %s)", runtime.GOOS, runtime.GOARCH, release.ThemeKitVersion.String()))
	}))

	client, _ = NewClient(Params{
		Domain:   server.URL,
		Password: "secret_password",
	})
	client.baseURL.Scheme = "http"

	resp, err = client.Get("/assets.json", map[string]string{"X-Custom-Header": "Foo"})
	assert.Nil(t, err)
	assert.NotNil(t, resp)

	resp, err = client.Delete("/assets.json", map[string]string{"X-Custom-Header": "Foo"})
	assert.Nil(t, err)
	assert.NotNil(t, resp)

	server.Close()

	var request uint64
	var mut sync.Mutex
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mut.Lock()
		if request == 0 {
			request++
			mut.Unlock()
			time.Sleep(5 * time.Millisecond)
		} else {
			mut.Unlock()
		}
	}))

	client, _ = NewClient(Params{
		Domain:  server.URL,
		Timeout: 5 * time.Millisecond,
	})
	client.baseURL.Scheme = "http"

	assert.NotNil(t, client)
	assert.Nil(t, err)

	_, err = client.do("POST", "/assets.json", body, nil)
	assert.Nil(t, err)
	server.Close()

	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(3 * time.Second)
	}))

	client, _ = NewClient(Params{
		Domain:  server.URL,
		Timeout: 5 * time.Millisecond,
	})
	client.maxRetry = 1
	client.baseURL.Scheme = "http"

	assert.NotNil(t, client)
	assert.Nil(t, err)

	_, err = client.do("POST", "/assets.json", body, nil)
	assert.Contains(t, err.Error(), "request failed after 1 retries", server.URL)
	server.Close()

	// Client should query Theme Access server instead of Shopify when password starts with a prefix "shptka_"
	shopifyServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))

	themeKitAccessServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, r.Header.Get("X-Shopify-Shop"), client.domain)
	}))

	client, err = NewClient(Params{
		Domain:   shopifyServer.URL,
		Password: "shptka_00000000000000000000000000000000",
	})
	themeKitAccessURL = themeKitAccessServer.URL

	assert.NotNil(t, client)
	assert.Nil(t, err)

	resp, err = client.Post("/assets.json", body, map[string]string{"X-Custom-Header": "Checksum"})
	assert.Nil(t, err)
	assert.NotNil(t, resp)

	server.Close()

	// Client should query Shopify instead of Theme Access server when password has no specified prefix
	shopifyServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Empty(t, r.Header.Get("X-Shopify-Shop"))
	}))

	themeKitAccessServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))

	client, err = NewClient(Params{
		Domain:   shopifyServer.URL,
		Password: "secret_password",
	})
	themeKitAccessURL = themeKitAccessServer.URL

	assert.NotNil(t, client)
	assert.Nil(t, err)

	resp, err = client.Post("/assets.json", body, map[string]string{"X-Custom-Header": "Checksum"})
	assert.Nil(t, err)
	assert.NotNil(t, resp)

	server.Close()
}

func TestGenerateHTTPAdapter(t *testing.T) {
	NewClient(Params{
		Domain:  "https://shop.myshopify.com",
		Timeout: 60 * time.Second,
	})
	assert.Equal(t, httpClient.Timeout, 60*time.Second)

	NewClient(Params{Domain: "https://shop.myshopify.com"})
	assert.Equal(t, httpClient.Timeout, 60*time.Second)
}

func TestProxyConfig(t *testing.T) {
	testcases := []struct {
		proxyURL, err string
	}{
		{proxyURL: ""},
		{proxyURL: "http//localhost:3000", err: "invalid proxy URI"},
		{proxyURL: "http://127.0.0.1:8080"},
	}

	for _, testcase := range testcases {
		_, err := NewClient(Params{
			Domain: "https://shop.myshopify.com",
			Proxy:  testcase.proxyURL,
		})
		if testcase.err == "" && assert.Nil(t, err) {
			if testcase.proxyURL == "" {
				assert.Nil(t, httpTransport.Proxy)
			} else {
				assert.NotNil(t, httpTransport.Proxy)
			}
		} else if assert.NotNil(t, err) {
			assert.Contains(t, err.Error(), testcase.err)
		}
	}
}

func TestParseBaseUrl(t *testing.T) {
	testcases := []struct {
		domain, expected, err string
	}{
		{domain: "test.myshopify.com", expected: "https://test.myshopify.com"},
		{domain: "$%@#.myshopify.com", expected: "", err: "invalid domain"},
	}

	for _, testcase := range testcases {
		actual, err := parseBaseURL(testcase.domain)
		if testcase.err == "" && assert.Nil(t, err) {
			assert.Equal(t, actual.String(), testcase.expected)
		} else if assert.NotNil(t, err) {
			assert.Contains(t, err.Error(), testcase.err)
		}
	}
}
