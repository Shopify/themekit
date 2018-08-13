package httpify

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"runtime"
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
		APILimit: time.Nanosecond,
	})
	client.baseURL.Scheme = "http"

	assert.NotNil(t, client)
	assert.Nil(t, err)

	resp, err := client.Post("/assets.json", body)
	assert.Nil(t, err)
	assert.NotNil(t, resp)

	resp, err = client.Put("/assets.json", body)
	assert.Nil(t, err)
	assert.NotNil(t, resp)

	server.Close()

	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, r.Header.Get("X-Shopify-Access-Token"), client.password)
		assert.Equal(t, r.Header.Get("Content-Type"), "application/json")
		assert.Equal(t, r.Header.Get("Accept"), "application/json")
		assert.Equal(t, r.Header.Get("User-Agent"), fmt.Sprintf("go/themekit (%s; %s; %s)", runtime.GOOS, runtime.GOARCH, release.ThemeKitVersion.String()))
	}))

	client, err = NewClient(Params{
		Domain:   server.URL,
		Password: "secret_password",
		APILimit: time.Nanosecond,
	})
	client.baseURL.Scheme = "http"

	resp, err = client.Get("/assets.json")
	assert.Nil(t, err)
	assert.NotNil(t, resp)

	resp, err = client.Delete("/assets.json")
	assert.Nil(t, err)
	assert.NotNil(t, resp)

	server.Close()

	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(3 * time.Second)
	}))
	defer server.Close()

	client, _ = NewClient(Params{
		Domain:   server.URL,
		Timeout:  time.Nanosecond,
		APILimit: time.Nanosecond,
	})
	client.baseURL.Scheme = "http"

	assert.NotNil(t, client)
	assert.Nil(t, err)

	_, err = client.do("POST", "/assets.json", body)
	if assert.NotNil(t, err) {
		assert.Equal(t, err, errClientTimeout)
	}
}

func TestGenerateHTTPAdapter(t *testing.T) {
	_, err := generateHTTPAdapter(time.Second, "#$#$^$%^##$")
	if assert.NotNil(t, err) {
		assert.EqualError(t, err, "invalid proxy URI")
	}

	c, err := generateHTTPAdapter(time.Second, "http://localhost:3000")
	assert.Nil(t, err)
	assert.Equal(t, time.Second, c.Timeout)
	assert.NotNil(t, c.Transport)
}

func TestGenerateClientTransport(t *testing.T) {
	testcases := []struct {
		proxyURL, err string
		expectNil     bool
	}{
		{proxyURL: "", expectNil: true},
		{proxyURL: "http//localhost:3000", expectNil: true, err: "invalid proxy URI"},
		{proxyURL: "http://127.0.0.1:8080", expectNil: false},
	}

	for _, testcase := range testcases {
		transport, err := generateClientTransport(testcase.proxyURL)
		assert.Equal(t, transport == nil, testcase.expectNil)
		if testcase.err == "" {
			assert.Nil(t, err)
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
