package kittest

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"sync"

	"github.com/ryanuber/go-glob"
)

var (
	// FixturesPath is the path that any files make during testing will be put under
	FixturesPath = filepath.Join(".", "fixtures")
	// FixtureProjectPath is a directory where any projects files will be put
	FixtureProjectPath = filepath.Join(FixturesPath, "project")
)

type (
	// TestRequest encapsulates a request made to the test server
	TestRequest struct {
		Method string
		URL    *url.URL
	}
	// Server is a test server that will record requests and server shopify responses
	Server struct {
		*httptest.Server
		sync.Mutex
		Requests []TestRequest
	}
)

func (server *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	server.Lock()
	defer server.Unlock()
	server.Requests = append(server.Requests, TestRequest{
		Method: r.Method,
		URL:    r.URL,
	})
	if r.URL.Path == "/themekit_update" {
		fmt.Fprintf(w, themekitUpdateFeed)
	} else if r.URL.Path == "/feed" {
		fmt.Fprintf(w, releaseAtom)
	} else if r.URL.Path == "/admin/themes.json" {
		fmt.Fprintf(w, themesResponse)
	} else if glob.Glob("/admin/themes/*/assets.json", r.URL.Path) {
		if r.Method == "GET" && glob.Glob("*asset%5Bkey%5D=*", r.URL.RawQuery) {
			if glob.Glob("*nope*", r.URL.RawQuery) {
				w.WriteHeader(404)
				fmt.Fprintf(w, "not found")
			} else {
				fmt.Fprintf(w, assetResponse)
			}
		} else if r.Method != "GET" {
			decoder := json.NewDecoder(r.Body)
			var theme map[string]struct {
				Key string
			}
			decoder.Decode(&theme)
			defer r.Body.Close()

			if asset, ok := theme["asset"]; ok && asset.Key == "nope" {
				w.WriteHeader(409)
				fmt.Fprintf(w, "not found")
			} else if ok && asset.Key == "empty" {
				fmt.Fprintf(w, "{\"asset\": {\"key\":\"\"}")
			} else {
				fmt.Fprintf(w, assetResponse)
			}
		} else {
			fmt.Fprintf(w, assetsReponse)
		}
	} else if glob.Glob("/admin/themes/*.json", r.URL.Path) {
		fmt.Fprintf(w, themeResponse)
	} else {
		println("bad request", r.URL.Path)
		w.WriteHeader(404)
		fmt.Fprintf(w, "bad request")
	}
}

// Reset will reset the request log
func (server *Server) Reset() {
	server.Lock()
	defer server.Unlock()
	server.Requests = []TestRequest{}
}

// NewTestServer is test server setup to act like the shopify API so that we can
// test against it for different responses
func NewTestServer() *Server {
	testServer := &Server{
		Requests: []TestRequest{},
	}
	testServer.Server = httptest.NewServer(testServer)
	return testServer
}
