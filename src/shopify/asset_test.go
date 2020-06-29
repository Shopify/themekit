package shopify

import (
	"encoding/base64"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Shopify/themekit/src/env"
)

func TestFindAssets(t *testing.T) {
	goodEnv := &env.Env{Directory: filepath.Join("_testdata", "project")}
	badEnv := &env.Env{Directory: "nope"}

	testcases := []struct {
		e      *env.Env
		inputs []string
		err    string
		count  int
	}{
		{e: goodEnv, inputs: []string{filepath.Join("assets", "application.js")}, count: 1},
		{e: goodEnv, count: 10},
		{e: badEnv, count: 7, err: " "},
		{e: goodEnv, inputs: []string{"assets", "config/settings_data.json"}, count: 6},
		{e: goodEnv, inputs: []string{"snippets/nope.txt"}, err: "readAsset: "},
	}

	for _, testcase := range testcases {
		assets, err := FindAssets(testcase.e, testcase.inputs...)
		if testcase.err == "" {
			assert.Nil(t, err)
			assert.Equal(t, testcase.count, len(assets))
		} else if assert.NotNil(t, err) {
			assert.Contains(t, err.Error(), testcase.err)
		}
	}
}

func TestAsset_Write(t *testing.T) {
	testDir := filepath.Join("_testdata", "writeto")
	os.Mkdir(testDir, 0755)

	testcases := []struct {
		filename, outdir, err string
	}{
		{outdir: testDir, filename: "blah.txt"},
		{outdir: "nothere", filename: "blah.txt", err: " "},
		{outdir: testDir, filename: filepath.Join("assets", "test.txt")},
	}

	for _, testcase := range testcases {
		err := Asset{Key: testcase.filename}.Write(testcase.outdir)
		if testcase.err == "" {
			assert.Nil(t, err)
			_, err := os.Stat(filepath.Join(testDir, testcase.filename))
			assert.Nil(t, err)
		} else if assert.NotNil(t, err) {
			assert.Contains(t, err.Error(), testcase.err)
		}
	}

	os.RemoveAll(testDir)
}

func TestAsset_Contents(t *testing.T) {
	testcases := []struct {
		asset  Asset
		err    string
		length int
	}{
		{asset: Asset{Value: "this is content"}, length: 15},
		{asset: Asset{Attachment: "this is bad content"}, err: "Could not decode"},
		{asset: Asset{Attachment: base64.StdEncoding.EncodeToString([]byte("this is good content"))}, length: 20},
		{asset: Asset{Key: "test.json", Value: "{\"test\":\"one\"}"}, length: 19},
	}

	for _, testcase := range testcases {
		data, err := testcase.asset.contents()
		if testcase.err == "" {
			assert.Nil(t, err)
			assert.Equal(t, testcase.length, len(data))
		} else if assert.NotNil(t, err) {
			assert.Contains(t, err.Error(), testcase.err)
		}
	}
}

func TestLoadAssetsFromDirectory(t *testing.T) {
	ignoreNone := func(path string) bool { return strings.Contains(path, ".gitkeep") }
	selectOne := func(path string) bool { return path != "assets/application.js" }

	testcases := []struct {
		path, err string
		ignore    func(string) bool
		count     int
	}{
		{path: "", ignore: ignoreNone, count: 10},
		{path: "", ignore: selectOne, count: 1},
		{path: "assets", ignore: ignoreNone, count: 5},
		{path: "nope", ignore: ignoreNone, count: 0, err: " "},
	}

	e := &env.Env{Directory: filepath.Join("_testdata", "project")}
	for _, testcase := range testcases {
		assets, err := loadAssetsFromDirectory(e, testcase.path, testcase.ignore)
		if testcase.err == "" {
			assert.Nil(t, err)
			assert.Equal(t, testcase.count, len(assets))
		} else if assert.NotNil(t, err) {
			assert.Contains(t, err.Error(), testcase.err)
		}
	}
}

func TestReadAsset(t *testing.T) {
	e := &env.Env{Directory: filepath.Join("_testdata", "project")}

	testcases := []struct {
		input    string
		expected Asset
		err      string
	}{
		{input: filepath.Join("assets", "application.js"), expected: Asset{Key: "assets/application.js", Value: "this is js content", Checksum: "f980fcdcfeb5bcf24c0de5c199c3a94b"}},
		{input: filepath.Join(".", "assets", "application.js"), expected: Asset{Key: "assets/application.js", Value: "this is js content", Checksum: "f980fcdcfeb5bcf24c0de5c199c3a94b"}},
		{input: "nope.txt", expected: Asset{}, err: " "},
		{input: "assets", expected: Asset{}, err: ErrAssetIsDir.Error()},
		{input: filepath.Join("assets", "image.png"), expected: Asset{Key: "assets/image.png", Attachment: "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAAEUlEQVR4nGJiYGBgAAQAAP//AA8AA/6P688AAAAASUVORK5CYII=", Checksum: "9e24e19b024c44b778301d880bd8e6f4"}},
		{input: filepath.Join("assets", "app.json"), expected: Asset{Key: "assets/app.json", Value: "{\"testing\" : \"data\"}", Checksum: "31409bedd9f5852166c0a4a9b874f1a7"}},
		// `app_alternate.json` has the same content but different whitespace. Since we normalise JSON before persisting, it has the same checksum.
		{input: filepath.Join("assets", "app_alternate.json"), expected: Asset{Key: "assets/app_alternate.json", Value: "{\"testing\":\"data\"}", Checksum: "31409bedd9f5852166c0a4a9b874f1a7"}},
		{input: filepath.Join("assets", "template.liquid"), expected: Asset{Key: "assets/template.liquid", Value: "{% comment %}\n  The contents\n{% endcomment %}\n", Checksum: "8bed50f07f1e8d52b07300e983c21a86"}},
	}

	for _, testcase := range testcases {
		actual, err := ReadAsset(e, testcase.input)
		if testcase.err == "" {
			assert.Nil(t, err)
			assert.Equal(t, testcase.expected.Key, actual.Key)
			assert.Contains(t, actual.Value, testcase.expected.Value) // contains because of line endings
			assert.Equal(t, testcase.expected.Checksum, actual.Checksum)
		} else if assert.NotNil(t, err) {
			assert.Contains(t, err.Error(), testcase.err)
		}
	}
}
