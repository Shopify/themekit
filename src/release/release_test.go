package release

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/Shopify/themekit/src/release/_mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestRelease_IsValid(t *testing.T) {
	r := release{}
	assert.False(t, r.isValid())

	r.Platforms = []platform{{Name: "test"}}
	assert.True(t, r.isValid())
}

func TestRelease_IsApplicable(t *testing.T) {
	testcases := []struct {
		in         string
		applicable bool
	}{
		{"20.0.0", true}, {"0.0.0", false}, {ThemeKitVersion.String(), false},
		{"v2.7.5-beta", false}, {"v2.7.5+prerelease", false}, {"this_is_not_a_version", false},
	}

	for _, testcase := range testcases {
		assert.Equal(t, release{Version: testcase.in}.isApplicable(), testcase.applicable)
	}
}

func TestRelease_GetVersion(t *testing.T) {
	assert.Equal(t, "20.0.0", release{Version: "20.0.0"}.getVersion().String())
}

func TestRelease_ForCurrentPlatform(t *testing.T) {
	thisSystem := runtime.GOOS + "-" + runtime.GOARCH

	r := release{Version: "20.0.0", Platforms: []platform{
		{Name: thisSystem},
		{Name: "other-system"},
		{Name: "bad-system"},
	}}
	assert.Equal(t, thisSystem, r.forCurrentPlatform().Name)

	r = release{Version: "20.0.0", Platforms: []platform{
		{Name: "other-system"},
		{Name: "bad-system"},
	}}
	assert.Equal(t, "", r.forCurrentPlatform().Name)
}

func TestIsUpdateAvailable(t *testing.T) {
	testcases := []struct {
		in         string
		applicable bool
	}{
		{"20.0.0", true}, {"0.0.0", false}, {ThemeKitVersion.String(), false},
		{"v2.7.5-beta", false}, {"v2.7.5+prerelease", false}, {"this_is_not_a_version", false},
	}

	for _, testcase := range testcases {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintln(w, `{"version":"`+testcase.in+`", "platforms": [{}]}`)
		}))
		assert.Equal(t, checkUpdateAvailable(ts.URL), testcase.applicable)
		ts.Close()
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	ts.Close()
	assert.False(t, checkUpdateAvailable(ts.URL))
}

func TestInstallLatest(t *testing.T) {
	testcases := []struct {
		in, err string
	}{
		{"20.0.0", ""}, {"0.0.0", "no applicable update available"}, {ThemeKitVersion.String(), "no applicable update available"},
	}

	for _, testcase := range testcases {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			name := runtime.GOOS + "-" + runtime.GOARCH
			fmt.Fprintln(w, `{"version":"`+testcase.in+`", "platforms": [{"name": "`+name+`"}]}`)
		}))
		err := installLatest(ts.URL, func(p platform) error { return nil })
		if testcase.err == "" {
			assert.Nil(t, err)
		} else if assert.NotNil(t, err) {
			assert.Contains(t, err.Error(), testcase.err)
		}
		ts.Close()
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	ts.Close()
	assert.Error(t, installLatest(ts.URL, func(p platform) error { return nil }))
}

func TestInstallVersion(t *testing.T) {
	testcases := []struct {
		in, err string
	}{
		{in: "0.4.7"},
		{in: "0.0.0", err: "version 0.0.0 not found"},
		{in: "v.0.8.2", err: "Malformed version: v.0.8.2"},
	}

	for _, testcase := range testcases {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintln(w, `[{"version":"0.4.7", "platforms": [{"name": "plat-name"}]}]`)
		}))
		err := installVersion(testcase.in, ts.URL, func(p platform) error { return nil })
		if testcase.err == "" {
			assert.Nil(t, err)
		} else if assert.NotNil(t, err) {
			assert.Contains(t, err.Error(), testcase.err)
		}
		ts.Close()
		if testcase.err == "" && assert.Nil(t, err) {
			err = installVersion(testcase.in, ts.URL, func(p platform) error { return nil })
			assert.NotNil(t, err)
		}
	}
}

func TestApplyUpdate(t *testing.T) {
	updateFile := []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06}
	sum := md5.Sum(updateFile)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { fmt.Fprintf(w, string(updateFile)) }))
	p := platform{URL: ts.URL, Digest: hex.EncodeToString(sum[:])}
	installtoDir := filepath.Join("_testdata", "installto")
	installto := filepath.Join(installtoDir, "updateme")
	assert.Nil(t, os.Mkdir(installtoDir, 0755))
	updateme, err := os.Create(installto)
	assert.Nil(t, err)
	updateme.Close()

	assert.Nil(t, applyUpdate(p, installto))
	buf, err := ioutil.ReadFile(installto)
	assert.Nil(t, err)
	assert.Equal(t, updateFile, buf)

	assert.NotNil(t, applyUpdate(p, filepath.Join("_testdata", "no", "updateme")))
	assert.NotNil(t, applyUpdate(platform{}, installto))
	assert.NotNil(t, applyUpdate(platform{Digest: "abcde"}, installto))

	ts.Close()
	assert.NotNil(t, applyUpdate(platform{URL: ts.URL}, installto))
	assert.Nil(t, os.RemoveAll(installtoDir))
}

func TestUpdate(t *testing.T) {
	testcases := []struct {
		ver, dir, err string
		force, req    bool
	}{
		{ver: "12.34.56", err: "deploy version does not match themekit version"},
		{ver: ThemeKitVersion.String(), dir: filepath.Join("_testdata", "nope"), err: "Dist folder does not exist"},
		{ver: ThemeKitVersion.String(), dir: filepath.Join("_testdata", "dist"), req: true, err: "version has already been deployed"},
		{ver: ThemeKitVersion.String(), dir: filepath.Join("_testdata", "otherdist"), err: " "},
		{ver: ThemeKitVersion.String(), dir: filepath.Join("_testdata", "dist")},
		{ver: "v.0.8.2", err: "Malformed version: v.0.8.2"},
	}

	for _, testcase := range testcases {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if testcase.req {
				fmt.Fprintf(w, `[{"version":"`+ThemeKitVersion.String()+`", "platforms": [{"name": "plat-name"}]}]`)
			} else {
				fmt.Fprintf(w, `[{"version":"0.4.7", "platforms": [{"name": "plat-name"}]}]`)
			}
		}))

		err := update(testcase.ver, ts.URL, testcase.dir, testcase.force, new(mocks.LaxUploader))
		if testcase.err == "" {
			assert.Nil(t, err)
		} else if assert.NotNil(t, err) {
			assert.Contains(t, err.Error(), testcase.err)
		}

		ts.Close()
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	ts.Close()
	err := update(ThemeKitVersion.String(), ts.URL, filepath.Join("_testdata", "dist"), false, new(mocks.LaxUploader))
	assert.Error(t, err)
}

func TestRemove(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `[{"version":"0.4.7", "platforms": [{"name": "plat-name"}]}]`)
	}))

	testcases := []struct {
		ver, dir, err string
		force, req    bool
	}{
		{ver: "12.34.56", err: "version has not be deployed"},
		{ver: "0.4.7"},
		{ver: "v.0.8.2", err: "Malformed version: v.0.8.2"},
	}

	for _, testcase := range testcases {
		err := remove(testcase.ver, ts.URL, new(mocks.LaxUploader))
		if testcase.err == "" {
			assert.Nil(t, err)
		} else if assert.NotNil(t, err) {
			assert.Contains(t, err.Error(), testcase.err)
		}
	}

	ts.Close()
	assert.Error(t, remove(ThemeKitVersion.String(), ts.URL, new(mocks.LaxUploader)))
}

func TestUpdateDeploy(t *testing.T) {
	latest := release{Version: "0.4.7"}
	releases := releasesList{release{Version: "0.4.4"}, latest}
	m := new(mocks.Uploader)
	m.On("JSON", "releases/all.json", releases).Return(nil)
	m.On("JSON", "releases/latest.json", latest).Return(nil)
	updateDeploy(releases, m)
	m.AssertExpectations(t)

	m = new(mocks.Uploader)
	m.On("JSON", "releases/all.json", releases).Return(errors.New("nope"))
	assert.Error(t, updateDeploy(releases, m))
}

func TestFetchLatest(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/not_json" {
			fmt.Fprintf(w, "this is not json")
		} else if r.URL.Path == "/good" {
			fmt.Fprintln(w, `{"version":"0.4.7"}`)
		} else {
			w.WriteHeader(404)
			fmt.Fprintf(w, "bad request")
		}
	}))
	defer ts.Close()
	r, err := fetchLatest(ts.URL + "/good")
	assert.Nil(t, err)
	assert.Equal(t, r, release{Version: "0.4.7"})
	_, err = fetchLatest(ts.URL + "/not_json")
	assert.NotNil(t, err)
	_, err = fetchLatest(ts.URL + "/doesntexist")
	assert.NotNil(t, err)
	ts.Close()
	_, err = fetchLatest(ts.URL)
	assert.NotNil(t, err)
}

func TestBuildRelease(t *testing.T) {
	u := new(mocks.Uploader)
	_, err := buildRelease("0.4.7", filepath.Join("_testdata", "otherdist"), u)
	assert.NotNil(t, err)

	u.On(
		"File",
		mock.MatchedBy(func(string) bool { return true }),
		mock.MatchedBy(func(*os.File) bool { return true }),
	).Return("http://amazon/themekit", nil)

	r, _ := buildRelease("0.4.7", filepath.Join("_testdata", "dist"), u)
	u.AssertExpectations(t)
	assert.Equal(t, r.Version, "0.4.7")

	testplatforms := []platform{
		{Name: "darwin-amd64", URL: "http://amazon/themekit", Digest: "d41d8cd98f00b204e9800998ecf8427e"},
		{Name: "linux-386", URL: "http://amazon/themekit", Digest: "d41d8cd98f00b204e9800998ecf8427e"},
		{Name: "linux-amd64", URL: "http://amazon/themekit", Digest: "d41d8cd98f00b204e9800998ecf8427e"},
		{Name: "windows-386", URL: "http://amazon/themekit", Digest: "d41d8cd98f00b204e9800998ecf8427e"},
		{Name: "windows-amd64", URL: "http://amazon/themekit", Digest: "d41d8cd98f00b204e9800998ecf8427e"},
		{Name: "freebsd-386", URL: "http://amazon/themekit", Digest: "d41d8cd98f00b204e9800998ecf8427e"},
		{Name: "freebsd-amd64", URL: "http://amazon/themekit", Digest: "d41d8cd98f00b204e9800998ecf8427e"},
	}
	for _, p := range testplatforms {
		assert.Contains(t, r.Platforms, p)
	}

	u = new(mocks.Uploader)
	u.On(
		"File",
		mock.MatchedBy(func(string) bool { return true }),
		mock.MatchedBy(func(*os.File) bool { return true }),
	).Return("", errors.New("didnt work"))

	_, err = buildRelease("0.4.7", filepath.Join("_testdata", "dist"), u)
	assert.NotNil(t, err)
}
