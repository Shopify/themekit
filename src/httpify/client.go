package httpify

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"runtime"
	"strings"
	"time"

	"github.com/Shopify/themekit/src/ratelimiter"
	"github.com/Shopify/themekit/src/release"
	"github.com/Shopify/themekit/src/util"
)

var (
	// ErrConnectionIssue is an error that is thrown when a very specific error is
	// returned from our http request that usually implies bad connections.
	ErrConnectionIssue = errors.New("DNS problem while connecting to Shopify, this indicates a problem with your internet connection")
	// ErrInvalidProxyURL is returned if a proxy url has been passed but is improperly formatted
	ErrInvalidProxyURL = errors.New("invalid proxy URI")
	httpTransport      = &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	httpClient = &http.Client{
		Timeout: 30 * time.Second,
	}
	themeKitAccessURL = "https://theme-kit-access.shopifyapps.com/cli"
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
		httpClient.Transport = httpTransport
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
	appBaseURL := client.baseURL.String()

	// redirect to Theme Access
	if util.IsThemeAccessPassword(client.password) {
		appBaseURL = themeKitAccessURL
	}

	req, err := http.NewRequest(method, appBaseURL+path, nil)

	if err != nil {
		return nil, err
	}

	req.Header.Add("X-Shopify-Access-Token", client.password)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")
	req.Header.Add("User-Agent", fmt.Sprintf("go/themekit (%s; %s; %s)", runtime.GOOS, runtime.GOARCH, release.ThemeKitVersion.String()))
	if util.IsThemeAccessPassword(client.password) {
		req.Header.Add("X-Shopify-Shop", client.domain)
	}
	for label, value := range headers {
		req.Header.Add(label, value)
	}

	return client.doWithRetry(req, body)
}

func (client *HTTPClient) doWithRetry(req *http.Request, body interface{}) (*http.Response, error) {
	var (
		bodyData []byte
		resp     *http.Response
		err      error
	)

	if body != nil {
		bodyData, err = json.Marshal(body)
		if err != nil {
			return nil, err
		}
	}

	for attempt := 0; attempt <= client.maxRetry; attempt++ {
		resp, err = client.limit.GateReq(httpClient, req, bodyData)
		if err == nil && resp.StatusCode >= 100 && resp.StatusCode < 500 {
			return resp, nil
		} else if err != nil && strings.Contains(err.Error(), "no such host") {
			return nil, ErrConnectionIssue
		}
		time.Sleep(time.Duration(attempt) * time.Second)
	}

	return nil, fmt.Errorf("request failed after %v retries with error: %v", client.maxRetry, err)
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
