package httpify

import (
	"bytes"
	"context"
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
	// ErrConnectionIssue is an error that is thrown when a very specific error is
	// returned from our http request that usually implies bad connections.
	ErrConnectionIssue = errors.New("DNS problem while connecting to Shopify, this indicates a problem with your internet connection")
	// ErrInvalidProxyURL is returned if a proxy url has been passed but is improperly formatted
	ErrInvalidProxyURL = errors.New("invalid proxy URI")
	netDialer          = &net.Dialer{
		Timeout:   3 * time.Second,
		KeepAlive: 1 * time.Second,
	}
	httpTransport = &http.Transport{
		TLSClientConfig:       &tls.Config{InsecureSkipVerify: true},
		IdleConnTimeout:       time.Second,
		TLSHandshakeTimeout:   time.Second,
		ExpectContinueTimeout: time.Second,
		ResponseHeaderTimeout: time.Second,
		MaxIdleConnsPerHost:   10,
		DialContext: func(ctx context.Context, network, address string) (conn net.Conn, err error) {
			if conn, err = netDialer.DialContext(ctx, network, address); err != nil {
				return nil, err
			}
			deadline := time.Now().Add(5 * time.Second)
			conn.SetReadDeadline(deadline)
			return conn, conn.SetDeadline(deadline)
		},
	}
	httpClient = &http.Client{
		Transport: httpTransport,
		Timeout:   30 * time.Second,
	}
)

type proxyHandler func(*http.Request) (*url.URL, error)

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

	if params.Timeout != 0 {
		httpClient.Timeout = params.Timeout
	}

	if params.Proxy != "" {
		parsedURL, err := url.ParseRequestURI(params.Proxy)
		if err != nil {
			return nil, ErrInvalidProxyURL
		}
		httpTransport.Proxy = http.ProxyURL(parsedURL)
	}

	return &HTTPClient{
		domain:   params.Domain,
		password: params.Password,
		baseURL:  baseURL,
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
	for attempt := 0; attempt <= client.maxRetry; {
		// reset the body when non-nil for every request (rewind)
		if body != nil {
			data, err := json.Marshal(body)
			if err != nil {
				return nil, err
			}
			req.Body = ioutil.NopCloser(bytes.NewBuffer(data))
		}

		client.limit.Wait()
		resp, err := httpClient.Do(req)
		if err == nil {
			if resp.StatusCode >= 100 && resp.StatusCode <= 428 {
				return resp, nil
			} else if resp.StatusCode == http.StatusTooManyRequests {
				after, _ := strconv.ParseFloat(resp.Header.Get("Retry-After"), 10)
				client.limit.ResetAfter(time.Duration(after) * time.Second)
				continue
			}
		} else if strings.Contains(err.Error(), "no such host") {
			return nil, ErrConnectionIssue
		}
		attempt++
		time.Sleep(time.Duration(attempt) * time.Second)
	}
	return nil, fmt.Errorf("request failed after %v retries", client.maxRetry)
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
