package httpify

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"runtime"
	"strings"
	"time"

	"github.com/Shopify/themekit/src/ratelimiter"
	"github.com/Shopify/themekit/src/release"
)

var (
	errClientTimeout   = errors.New(`request timed out. if you are receive this error consistently, try increasing the timeout in your config`)
	errConnectionIssue = errors.New("DNS problem while connecting to Shopify, this indicates a problem with your internet connection")
)

// Params allows for a better structured input into NewClient
type Params struct {
	Domain   string
	Password string
	Proxy    string
	Timeout  time.Duration
	APILimit time.Duration
}

// HTTPClient encapsulates an authenticate http client to issue theme requests
// to Shopify
type HTTPClient struct {
	domain   string
	password string
	baseURL  *url.URL
	client   *http.Client
	limit    *ratelimiter.Limiter
}

// NewClient will create a new authenticated http client that will communicate
// with Shopify
func NewClient(params Params) (*HTTPClient, error) {
	baseURL, err := parseBaseURL(params.Domain)
	if err != nil {
		return nil, err
	}

	adapter, err := generateHTTPAdapter(params.Timeout, params.Proxy)
	if err != nil {
		return nil, err
	}

	return &HTTPClient{
		domain:   params.Domain,
		password: params.Password,
		baseURL:  baseURL,
		client:   adapter,
		limit:    ratelimiter.New(params.Domain, params.APILimit),
	}, nil
}

// Get will send a get request to the path provided
func (client *HTTPClient) Get(path string) (*http.Response, error) {
	return client.do("GET", path, nil)
}

// Post will send a Post request to the path provided and set the post body as the
// object passed
func (client *HTTPClient) Post(path string, body interface{}) (*http.Response, error) {
	return client.do("POST", path, body)
}

// Put will send a Put request to the path provided and set the post body as the
// object passed
func (client *HTTPClient) Put(path string, body interface{}) (*http.Response, error) {
	return client.do("PUT", path, body)
}

// Delete will send a delete request to the path provided
func (client *HTTPClient) Delete(path string) (*http.Response, error) {
	return client.do("DELETE", path, nil)
}

// DoJSON will issue an authenticated json request to shopify.
func (client *HTTPClient) do(method, path string, body interface{}) (*http.Response, error) {
	var jsonData io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		jsonData = bytes.NewBuffer(data)
	}

	req, err := http.NewRequest(method, client.baseURL.String()+path, jsonData)
	if err != nil {
		return nil, err
	}

	req.Header.Add("X-Shopify-Access-Token", client.password)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")
	req.Header.Add("User-Agent", fmt.Sprintf("go/themekit (%s; %s; %s)", runtime.GOOS, runtime.GOARCH, release.ThemeKitVersion.String()))

	client.limit.Wait()
	resp, err := client.client.Do(req)
	if err, ok := err.(net.Error); ok && err.Timeout() {
		return nil, errClientTimeout
	} else if err != nil && strings.Contains(err.Error(), "no such host") {
		return nil, errConnectionIssue
	}
	return resp, err
}

func generateHTTPAdapter(timeout time.Duration, proxyURL string) (*http.Client, error) {
	adapter := &http.Client{Timeout: timeout}
	if transport, err := generateClientTransport(proxyURL); err != nil {
		return nil, err
	} else if transport != nil {
		adapter.Transport = transport
	}
	return adapter, nil
}

func generateClientTransport(proxyURL string) (*http.Transport, error) {
	if proxyURL == "" {
		return nil, nil
	}

	parsedURL, err := url.ParseRequestURI(proxyURL)
	if err != nil {
		return nil, fmt.Errorf("invalid proxy URI")
	}

	return &http.Transport{
		Proxy:           http.ProxyURL(parsedURL),
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}, nil
}

func parseBaseURL(domain string) (*url.URL, error) {
	u, err := url.Parse(domain)
	if err != nil {
		return nil, fmt.Errorf("invalid domain %s", domain)
	}
	if u.Hostname() != "127.0.0.1" { //unless we are testing locally
		u.Scheme = "https"
	}
	return u, nil
}
