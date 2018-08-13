package release

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFetchRelease(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/not_json" {
			fmt.Fprintf(w, "this is not json")
		} else if r.URL.Path == "/good" {
			fmt.Fprintln(w, `[{"version":"0.4.4"},{"version":"0.4.7"},{"version":"0.4.8-prerelease"}]`)
		} else {
			w.WriteHeader(404)
			fmt.Fprintf(w, "bad request")
		}
	}))
	defer ts.Close()
	releases, err := fetchReleases(ts.URL + "/good")
	assert.Nil(t, err)
	assert.Equal(t, 3, len(releases))
	_, err = fetchReleases(ts.URL + "/not_json")
	assert.NotNil(t, err)
	_, err = fetchReleases(ts.URL + "/doesntexist")
	assert.NotNil(t, err)
	ts.Close()
	_, err = fetchReleases(ts.URL)
	assert.NotNil(t, err)
}

func TestReleasesList_Get(t *testing.T) {
	releases := releasesList{release{Version: "0.4.4"}, release{Version: "0.4.7"}, release{Version: "0.4.8-prerelease"}}
	r := releases.get("latest")
	assert.Equal(t, "0.4.7", r.Version)

	r = releases.get("0.4.4")
	assert.Equal(t, "0.4.4", r.Version)

	r = releases.get("0.0.0")
	assert.Equal(t, "", r.Version)
	assert.Equal(t, false, r.isValid())
}

func TestReleasesList_Del(t *testing.T) {
	releases := releasesList{release{Version: "0.4.4"}, release{Version: "0.4.7"}}

	releases = releases.del("0.4.7")
	r := releases.get("0.4.7")
	assert.Equal(t, "", r.Version)
	assert.Equal(t, 1, len(releases))

	old := releases
	releases = releases.del("0.4.1")
	assert.Equal(t, 1, len(releases))
	assert.Equal(t, old, releases)
}

func TestReleasesList_Add(t *testing.T) {
	releases := releasesList{release{Version: "0.4.4"}, release{Version: "0.4.7"}}

	releases = releases.add(release{Version: "0.4.4"})
	assert.Equal(t, 2, len(releases))

	releases = releases.add(release{Version: "0.4.5"})
	assert.Equal(t, 3, len(releases))
}
