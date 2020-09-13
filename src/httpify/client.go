package httpify

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/Shopify/themekit/src/ratelimiter"
	"github.com/Shopify/themekit/src/release"
)

var (
	errConnectionIssue = errors.New("DNS problem while connecting to Shopify, this indicates a problem with your internet connection")
)

// Params allows for a better structured input into NewClient
type Params struct {
	Domain   string
	Password string
	Proxy    string
	Timeout  time.Duration
}

// HTTPClient encapsulates an authenticate http client to issue theme requests
// to Shopify
type HTTPClient struct {
	domain   string
	password string
	baseURL  *url.URL
	client   *http.Client
	limit    *ratelimiter.Limiter
	maxRetry int
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
		limit:    ratelimiter.New(params.Domain, 4),
		maxRetry: 5,
	}, nil
}

// Get will send a get request to the path provided
func (client *HTTPClient) Get(path string, headers map[string]string) (*http.Response, error) {
	return client.do("GET", path, nil, headers)
}

// Post will send a Post request to the path provided and set the post body as the
// object passed
func (client *HTTPClient) Post(path string, body interface{}, headers map[string]string) (*http.Response, error) {
	return client.do("POST", path, body, headers)
}

// Put will send a Put request to the path provided and set the post body as the
// object passed
func (client *HTTPClient) Put(path string, body interface{}, headers map[string]string) (*http.Response, error) {
	return client.do("PUT", path, body, headers)
}

// Delete will send a delete request to the path provided
func (client *HTTPClient) Delete(path string, headers map[string]string) (*http.Response, error) {
	return client.do("DELETE", path, nil, headers)
}

// do will issue an authenticated json request to shopify.
func (client *HTTPClient) do(method, path string, body interface{}, headers map[string]string) (*http.Response, error) {
	req, err := http.NewRequest(method, client.baseURL.String()+path, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("X-Shopify-Access-Token", client.password)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")
	req.Header.Add("User-Agent", fmt.Sprintf("go/themekit (%s; %s; %s)", runtime.GOOS, runtime.GOARCH, release.ThemeKitVersion.String()))
	for label, value := range headers {
		req.Header.Add(label, value)
	}

	return client.doWithRetry(req, body)
}

func (client *HTTPClient) doWithRetry(req *http.Request, body interface{}) (*http.Response, error) {
	attempt := 0
	for {
		// reset the body when non-nil for every request (rewind)
		if body != nil {
			data, err := json.Marshal(body)
			if err != nil {
				return nil, err
			}
			req.Body = ioutil.NopCloser(bytes.NewBuffer(data))
		}

		client.limit.Wait()
		resp, err := client.client.Do(req)
		if err == nil && resp.StatusCode >= 100 && resp.StatusCode <= 428 {
			return resp, nil
		} else if err, ok := err.(net.Error); ok && err.Timeout() {
			attempt++
			if attempt > client.maxRetry {
				return resp, fmt.Errorf("request timed out after %v retries, there may be an issue with your connection", client.maxRetry)
			}
			time.Sleep(time.Duration(attempt) * time.Second)
		} else if err == nil && resp.StatusCode == http.StatusTooManyRequests {
			after, _ := strconv.ParseFloat(resp.Header.Get("Retry-After"), 10)
			client.limit.ResetAfter(time.Duration(after))
		} else if err != nil && strings.Contains(err.Error(), "no such host") {
			return nil, errConnectionIssue
		}
	}
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
