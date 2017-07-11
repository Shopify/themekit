package kittest

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"

	"github.com/ryanuber/go-glob"
)

var (
	// FixturesPath is the path that any files make during testing will be put under
	FixturesPath = filepath.Join(".", "fixtures")
	// FixtureProjectPath is a directory where any projects files will be put
	FixtureProjectPath = filepath.Join(FixturesPath, "project")
)

// NewTestServer is test server setup to act like the shopify API so that we can
// test against it for different responses
func NewTestServer() *httptest.Server {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/feed" {
			fmt.Fprintf(w, releaseAtom)
		} else if r.URL.Path == "/admin/themes.json" {
			fmt.Fprintf(w, themesResponse)
		} else if glob.Glob("/admin/themes/*/assets.json", r.URL.Path) {
			if glob.Glob("*asset%5Bkey%5D=*", r.URL.RawQuery) {
				if glob.Glob("*nope*", r.URL.RawQuery) {
					w.WriteHeader(404)
					fmt.Fprintf(w, "not found")
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
	}))

	return server
}
