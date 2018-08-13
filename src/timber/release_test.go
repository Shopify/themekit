package timber

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Shopify/themekit/src/atom"
)

const releaseAtom = `<?xml version="1.0" encoding="UTF-8"?>
<feed xmlns="http://www.w3.org/2005/Atom" xmlns:media="http://search.yahoo.com/mrss/" xml:lang="en-US">
  <entry> <title>v2.0.2</title> </entry>
  <entry><title>v2.0.1</title></entry>
  <entry><title>v2.0.0</title></entry>
  <entry><title>v1.3.1</title></entry>
  <entry><title>v1.0.0</title></entry>
</feed>`

func TestGetVersionPath(t *testing.T) {
	testcases := []struct {
		version, path, err string
	}{
		{version: "master", path: "https://github.com/Shopify/Timber/archive/master.zip"},
		{version: "v1.3.1", path: "https://github.com/Shopify/Timber/archive/v1.3.1.zip"},
		{version: "latest", path: "https://github.com/Shopify/Timber/archive/v2.0.2.zip"},
		{version: "v1.3.0", err: "Available Versions Are"},
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, releaseAtom)
	}))
	defer ts.Close()

	for _, testcase := range testcases {
		path, err := getVersionPath(testcase.version, ts.URL+"/feed")
		if testcase.err == "" {
			assert.Nil(t, err)
			assert.Equal(t, path, testcase.path)
		} else if assert.NotNil(t, err, testcase.err) {
			assert.Contains(t, err.Error(), testcase.err)
		}
	}

	ts.Close()
	_, err := getVersionPath("latest", ts.URL+"/feed")
	assert.NotNil(t, err)
}

func TestDownloadThemeReleaseAtomFeed(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, releaseAtom)
	}))
	_, err := downloadThemeReleaseAtomFeed(ts.URL)
	ts.Close()
	assert.Nil(t, err)
	_, err = downloadThemeReleaseAtomFeed(ts.URL)
	assert.NotNil(t, err)

	ts = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "this is not an atom feed")
	}))
	defer ts.Close()
	_, err = downloadThemeReleaseAtomFeed(ts.URL)
	if assert.NotNil(t, err) {
		assert.Contains(t, err.Error(), "EOF")
	}
}

func TestFindThemeReleaseWith(t *testing.T) {
	feed, _ := atom.LoadFeed(strings.NewReader(releaseAtom))

	testcases := []struct {
		version, path, err string
	}{
		{version: "v1.3.1", path: "v1.3.1"},
		{version: "latest", path: "v2.0.2"},
		{version: "v1.3.0", err: "Available Versions Are"},
	}

	for _, testcase := range testcases {
		entry, err := findThemeReleaseWith(feed, testcase.version)
		if testcase.err == "" {
			assert.Nil(t, err)
			assert.Equal(t, entry.Title, testcase.path)
		} else if assert.NotNil(t, err, testcase.err) {
			assert.Contains(t, err.Error(), testcase.err)
		}
	}
}
