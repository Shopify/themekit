package kittest

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"runtime"

	"github.com/ryanuber/go-glob"
)

var (
	// FixturesPath is the path that any files make during testing will be put under
	FixturesPath = filepath.Join(".", "fixtures")
	// FixtureProjectPath is a directory where any projects files will be put
	FixtureProjectPath = filepath.Join(FixturesPath, "project")
	// UpdateFilePath is the file for testing updating themekit
	UpdateFilePath = filepath.Join(FixturesPath, "updateme")
	// NewUpdateFile is the file that is sent for updating themekit
	NewUpdateFile = []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06}
	// NewUpdateFileChecksum is the checksum for newupdatefile
	NewUpdateFileChecksum = md5.Sum(NewUpdateFile)
)

type (
	// Server is a test server that will record requests and server shopify responses
	Server struct {
		deletedCompiledFile bool
		themePreviewable    bool
		*httptest.Server
		Requests []*http.Request
	}
)

func (server *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	server.Requests = append(server.Requests, r)
	if r.URL.Path == "/themekit_update" {
		fmt.Fprintf(w, themekitUpdateFeed)
	} else if r.URL.Path == "/themekit_system_update" {
		type platform struct {
			Name       string `json:"name"`
			URL        string `json:"url"`
			Digest     string `json:"digest"`
			TargetPath string `json:"TargetPath"` // used for testing updating
		}
		out, _ := json.Marshal([]struct {
			Version   string     `json:"version"`
			Platforms []platform `json:"platforms"`
		}{
			{
				Version: "20.0.0",
				Platforms: []platform{
					{
						Name:       runtime.GOOS + "-" + runtime.GOARCH,
						URL:        server.URL + "/release_download",
						Digest:     hex.EncodeToString(NewUpdateFileChecksum[:]),
						TargetPath: UpdateFilePath,
					},
				},
			},
		})
		fmt.Fprintf(w, string(out))
	} else if r.URL.Path == "/release_download" {
		fmt.Fprintf(w, string(NewUpdateFile))
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
			} else if ok && asset.Key == "templates/template.html" && r.Method == "PUT" && !server.deletedCompiledFile {
				w.WriteHeader(422)
				fmt.Fprintf(w, `{"errors":{"asset":["Cannot overwrite generated asset"]}}`)
			} else if ok && asset.Key == "templates/template.html.liquid" && r.Method == "DELETE" && !server.deletedCompiledFile {
				server.deletedCompiledFile = true
				w.WriteHeader(200)
				fmt.Fprintf(w, assetResponse)
			} else {
				fmt.Fprintf(w, assetResponse)
			}
		} else {
			fmt.Fprintf(w, assetsReponse)
		}
	} else if glob.Glob("/admin/themes/*.json", r.URL.Path) {
		if r.Method == "GET" && !server.themePreviewable {
			fmt.Fprintf(w, `{"theme":{"name":"timberland","role":"unpublished","previewable":false,"source":"https://githubz.com/shopify/timberlands"}}`)
			server.themePreviewable = true
		} else if r.Method == "GET" {
			fmt.Fprintf(w, `{"theme":{"name":"timberland","role":"unpublished","previewable":true,"source":"https://githubz.com/shopify/timberlands"}}`)
		} else if r.Method == "POST" {
			decoder := json.NewDecoder(r.Body)
			var theme map[string]struct {
				Name string
			}
			decoder.Decode(&theme)
			defer r.Body.Close()

			if asset, ok := theme["theme"]; ok && asset.Name == "nope" {
				fmt.Fprintf(w, `{"errors":{"src":["is empty"]}}`)
			} else {
				fmt.Fprintf(w, `{"theme":{"name":"timberland","role":"unpublished","previewable":false, "source":"https://githubz.com/shopify/timberlands"}}`)
			}
		}
	} else if r.URL.Path == "/not_json" {
		fmt.Fprintf(w, "this is not json")
	} else {
		w.WriteHeader(404)
		fmt.Fprintf(w, "bad request")
	}
}

// Reset will reset the request log
func (server *Server) Reset() {
	server.Requests = []*http.Request{}
}

// NewTestServer is test server setup to act like the shopify API so that we can
// test against it for different responses
func NewTestServer() *Server {
	testServer := &Server{
		Requests: []*http.Request{},
	}
	testServer.Server = httptest.NewServer(testServer)
	return testServer
}
