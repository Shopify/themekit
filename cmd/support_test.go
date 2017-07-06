package cmd

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"

	"github.com/ryanuber/go-glob"

	"github.com/Shopify/themekit/cmd/internal/atom"
	"github.com/Shopify/themekit/kit"
)

var (
	fixturesPath         = filepath.Join("..", "fixtures")
	fixtureProjectPath   = filepath.Join(fixturesPath, "project")
	goodEnvirontmentPath = filepath.Join(fixtureProjectPath, "valid_config.yml")
	badEnvirontmentPath  = filepath.Join(fixtureProjectPath, "invalid_config.yml")
	releasesPath         = filepath.Join(fixturesPath, "releases.atom")
)

func resetArbiter() {
	arbiter = newCommandArbiter()
	arbiter.verbose = true
	arbiter.setFlagConfig()
}

func getClient() (kit.ThemeClient, error) {
	arbiter.configPath = goodEnvirontmentPath
	if err := arbiter.generateThemeClients(nil, []string{}); err != nil {
		return kit.ThemeClient{}, err
	}
	return arbiter.activeThemeClients[0], nil
}

func loadAtom() atom.Feed {
	stream, _ := os.Open("../fixtures/releases.atom")
	feed, _ := atom.LoadFeed(stream)
	return feed
}

func fileFixture(name string) *os.File {
	path := fmt.Sprintf("../fixtures/%s.json", name)
	file, _ := os.Open(path)
	return file
}

func jsonFixture(name string) string {
	bytes, err := ioutil.ReadAll(fileFixture(name))
	if err != nil {
		log.Fatal(err)
	}
	return string(bytes)
}

func newTestServer() *httptest.Server {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/feed" {
			file, _ := os.Open(releasesPath)
			bytes, _ := ioutil.ReadAll(file)
			fmt.Fprintf(w, string(bytes))
		} else if r.URL.Path == "/domain/admin/themes.json" {
			fmt.Fprintf(w, jsonFixture("responses/themes"))
		} else if glob.Glob("/domain/admin/themes/*.json", r.URL.Path) {
			fmt.Fprintf(w, jsonFixture("responses/theme"))
		} else if glob.Glob("/domain/admin/themes/*/assets.json", r.URL.Path) {
			if glob.Glob("asset[key]=*", r.URL.RawQuery) {
				fmt.Fprintf(w, jsonFixture("responses/asset"))
			} else {
				fmt.Fprintf(w, jsonFixture("responses/assets"))
			}
		} else {
			println("bad request", r.URL.Path)
			w.WriteHeader(404)
			fmt.Fprintf(w, "bad request")
		}
	}))

	timberFeedPath = server.URL + "/feed"
	themeZipRoot = server.URL + "/zip"

	return server
}
