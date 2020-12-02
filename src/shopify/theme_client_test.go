package shopify

import (
	"errors"
	"fmt"
	"net/http"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Shopify/themekit/src/env"
	"github.com/Shopify/themekit/src/shopify/_mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var NoHeaders = map[string]string(nil)

// APIPath is the version of the Admin REST API to use

func TestNewThemeClient(t *testing.T) {
	testcases := []struct {
		e   *env.Env
		err string
	}{
		{e: &env.Env{ThemeID: "test123"}},
		{e: &env.Env{Directory: filepath.Join("_testdata", "project"), Proxy: "://foo.com"}, err: "invalid proxy URI"},
		{e: &env.Env{Ignores: []string{"nope"}}, err: " "},
	}

	for _, testcase := range testcases {
		client, err := NewClient(testcase.e)
		if testcase.err == "" {
			assert.Nil(t, err)
			assert.Equal(t, client.themeID, testcase.e.ThemeID)
			assert.NotNil(t, client.http)
			assert.NotNil(t, client.filter)
		} else if assert.NotNil(t, err) {
			assert.Contains(t, err.Error(), testcase.err)
		}
	}
}

func TestThemeClient_GetShop(t *testing.T) {
	testcases := []struct {
		themeID, resp, resperr, err string
		code                        int
	}{
		{resp: `{"errors": "Not Found"}`, code: 200, err: "Not Found"},
		{resperr: "(Client.Timeout exceeded while awaiting headers)", err: "(Client.Timeout exceeded while awaiting headers)"},
		{resp: `{"id": 123456}`, code: 200},
		{resp: "{}", code: 404, err: ErrShopDomainNotFound.Error()},
	}

	for _, testcase := range testcases {
		m := new(mocks.HttpAdapter)
		client, _ := NewClient(&env.Env{ThemeID: testcase.themeID})
		client.http = m

		expectation := m.On("Get", "/meta.json", NoHeaders)
		if testcase.resperr != "" {
			expectation.Return(nil, errors.New(testcase.resperr))
		} else {
			expectation.Return(jsonResponse(testcase.resp, testcase.code), nil)
		}

		shop, err := client.GetShop()

		if testcase.err == "" {
			assert.Nil(t, err)
			assert.Equal(t, shop.ID, int64(123456))
		} else if assert.NotNil(t, err) {
			assert.Contains(t, err.Error(), testcase.err)
		}

		if testcase.resp != "" || testcase.resperr != "" {
			m.AssertExpectations(t)
		}
	}
}

func TestThemeClient_Themes(t *testing.T) {
	testcases := []struct {
		resp, resperr, err string
		code               int
	}{
		{resp: `{"errors": "Not Found"}`, code: 200, err: "Not Found"},
		{resperr: "(Client.Timeout exceeded while awaiting headers)", err: "(Client.Timeout exceeded while awaiting headers)"},
		{resp: `{"themes":[{"id": 123456}]}`, code: 200},
	}

	for _, testcase := range testcases {
		m := new(mocks.HttpAdapter)
		client, _ := NewClient(&env.Env{})
		client.http = m

		expectation := m.On("Get", APIPath+"themes.json", NoHeaders)
		if testcase.resperr != "" {
			expectation.Return(nil, errors.New(testcase.resperr))
		} else {
			expectation.Return(jsonResponse(testcase.resp, testcase.code), nil)
		}

		themes, err := client.Themes()

		if testcase.err == "" {
			assert.Nil(t, err)
			assert.Equal(t, themes[0].ID, int64(123456))
		} else if assert.NotNil(t, err) {
			assert.Contains(t, err.Error(), testcase.err)
		}

		if testcase.resp != "" || testcase.resperr != "" {
			m.AssertExpectations(t)
		}
	}
}

func TestThemeClient_CreateNewTheme(t *testing.T) {
	testcases := []struct {
		in                 string
		theme              Theme
		resp, resperr, err string
	}{
		{in: "", err: ErrThemeNameRequired.Error()},
		{in: "my theme", resp: `{"errors": "Not Found"}`, err: "Not Found"},
		{in: "my theme", resperr: "(Client.Timeout exceeded while awaiting headers)", err: "(Client.Timeout exceeded while awaiting headers)"},
		{in: "my theme", resp: `{"theme":{"id": 123456,"name":"timberland","role":"unpublished","previewable":false}}`},
	}

	for _, testcase := range testcases {
		client, _ := NewClient(&env.Env{})
		m := new(mocks.HttpAdapter)
		client.http = m
		query := map[string]interface{}{"theme": Theme{Name: testcase.in}}

		if testcase.resp != "" {
			m.On("Post", APIPath+"themes.json", query, NoHeaders).Return(jsonResponse(testcase.resp, 200), nil)
		} else if testcase.resperr != "" {
			m.On("Post", APIPath+"themes.json", query, NoHeaders).Return(nil, errors.New(testcase.resperr))
		}

		theme, err := client.CreateNewTheme(testcase.in)

		if testcase.err == "" {
			assert.Nil(t, err)
			assert.Equal(t, theme.ID, int64(123456))
			assert.Equal(t, theme.Name, "timberland")
		} else if assert.NotNil(t, err) {
			assert.Contains(t, err.Error(), testcase.err)
		}

		if testcase.resp != "" || testcase.resperr != "" {
			m.AssertExpectations(t)
		}
	}
}

func TestThemeClient_GetInfo(t *testing.T) {
	testcases := []struct {
		themeID, resp, resperr, err string
		code                        int
	}{
		{err: ErrInfoWithoutThemeID.Error()},
		{themeID: "nope", resp: `{"errors": "Not Found"}`, code: 200, err: "Not Found"},
		{themeID: "123456", resperr: "(Client.Timeout exceeded while awaiting headers)", err: "(Client.Timeout exceeded while awaiting headers)"},
		{themeID: "123456", resp: `{"theme":{"id": 123456,"name":"timberland","role":"unpublished","previewable":false}}`, code: 200},
		{themeID: "123456", resp: "{}", code: 404, err: ErrThemeNotFound.Error()},
	}

	for _, testcase := range testcases {
		m := new(mocks.HttpAdapter)
		client, _ := NewClient(&env.Env{ThemeID: testcase.themeID})
		client.http = m

		expectation := m.On("Get", fmt.Sprintf(APIPath+"themes/%s.json", testcase.themeID), NoHeaders)
		if testcase.resperr != "" {
			expectation.Return(nil, errors.New(testcase.resperr))
		} else {
			expectation.Return(jsonResponse(testcase.resp, testcase.code), nil)
		}

		theme, err := client.GetInfo()

		if testcase.err == "" {
			assert.Nil(t, err)
			assert.Equal(t, theme.ID, int64(123456))
			assert.Equal(t, theme.Name, "timberland")
		} else if assert.NotNil(t, err) {
			assert.Contains(t, err.Error(), testcase.err)
		}

		if testcase.resp != "" || testcase.resperr != "" {
			m.AssertExpectations(t)
		}
	}
}

func TestThemeClient_PublishTheme(t *testing.T) {
	testcases := []struct {
		themeID, resp, resperr, err string
		code                        int
	}{
		{err: ErrPublishWithoutThemeID.Error()},
		{themeID: "nope", resp: `{"errors": "Not Found"}`, code: 200, err: "Not Found"},
		{themeID: "123456", resperr: "(Client.Timeout exceeded while awaiting headers)", err: "(Client.Timeout exceeded while awaiting headers)"},
		{themeID: "123456", resp: `{"theme":{"id": 123456,"name":"timberland","role":"main","previewable":true}}`, code: 200},
		{themeID: "123456", resp: `{"errors":{"role":["cannot be set to main: missing required file layout/theme.liquid"]}}`, code: 422, err: "role cannot be set to main: missing required file layout/theme.liquid"},
		{themeID: "123456", resp: "{}", code: 404, err: ErrThemeNotFound.Error()},
	}

	for i, testcase := range testcases {
		m := new(mocks.HttpAdapter)
		client, _ := NewClient(&env.Env{ThemeID: testcase.themeID})
		client.http = m

		expectation := m.On(
			"Put",
			fmt.Sprintf(APIPath+"themes/%s.json", testcase.themeID),
			map[string]Theme{"theme": {Role: "main"}},
			NoHeaders,
		)
		if testcase.resperr != "" {
			expectation.Return(nil, errors.New(testcase.resperr))
		} else {
			expectation.Return(jsonResponse(testcase.resp, testcase.code), nil)
		}

		err := client.PublishTheme()

		if testcase.err == "" {
			assert.Nil(t, err, fmt.Sprintf("unexpected err in testcase: %d", i))
		} else if assert.NotNil(t, err) {
			assert.Contains(t, err.Error(), testcase.err)
		}

		if testcase.resp != "" || testcase.resperr != "" {
			m.AssertExpectations(t)
		}
	}
}

func TestThemeClient_GetAllAssets(t *testing.T) {
	testcases := []struct {
		resp, resperr, err string
		code               int
	}{
		{resp: `{"errors": "Not Found"}`, code: 200, err: "Not Found"},
		{resperr: "(Client.Timeout exceeded while awaiting headers)", err: "(Client.Timeout exceeded while awaiting headers)"},
		{resp: `{"assets":[{"key":"assets/hello.txt"},{"key":"assets/goodbye.txt"}]}`, code: 200},
		{resp: "{}", code: 404, err: ErrThemeNotFound.Error()},
	}

	for _, testcase := range testcases {
		m := new(mocks.HttpAdapter)
		client, _ := NewClient(&env.Env{ThemeID: "123"})
		client.http = m

		expectation := m.On("Get", APIPath+"themes/123/assets.json?fields=key%2Cchecksum", NoHeaders)
		if testcase.resperr != "" {
			expectation.Return(nil, errors.New(testcase.resperr))
		} else {
			expectation.Return(jsonResponse(testcase.resp, testcase.code), nil)
		}

		assets, err := client.GetAllAssets()

		if testcase.err == "" {
			assert.Nil(t, err)
			assert.True(t, len(assets) > 0)
		} else if assert.NotNil(t, err, testcase.err) {
			assert.Contains(t, err.Error(), testcase.err)
		}

		m.AssertExpectations(t)
	}

	filtertestcases := []struct {
		input    string
		ignore   []string
		expected []Asset
	}{
		{
			input:    `{"assets":[{"key":"templates/foo.json.liquid"},{"key":"templates/foo.json"}]}`,
			expected: []Asset{{Key: "templates/foo.json.liquid"}},
		},
		{
			input:    `{"assets":[{"key":"templates/foo.json"},{"key":"templates/foo.json.liquid"}]}`,
			expected: []Asset{{Key: "templates/foo.json.liquid"}},
		},
		{
			input:    `{"assets":[{"key":"templates/ignore.html.liquid"},{"key":"templates/other.liquid"}]}`,
			expected: []Asset{{Key: "templates/other.liquid"}},
			ignore:   []string{"templates/ignore.html.liquid"},
		},
	}

	for _, testcase := range filtertestcases {
		m := new(mocks.HttpAdapter)
		client, _ := NewClient(&env.Env{ThemeID: "123", IgnoredFiles: testcase.ignore})
		client.http = m
		m.On("Get", APIPath+"themes/123/assets.json?fields=key%2Cchecksum", NoHeaders).Return(jsonResponse(testcase.input, 200), nil)
		assets, err := client.GetAllAssets()
		assert.Nil(t, err)
		assert.Equal(t, testcase.expected, assets)
	}
}

func TestThemeClient_GetAsset(t *testing.T) {
	testcases := []struct {
		resp, resperr, err string
		code               int
	}{
		{resp: `{"errors": "Not Found"}`, code: 200, err: "Not Found"},
		{resperr: "(Client.Timeout exceeded while awaiting headers)", err: "(Client.Timeout exceeded while awaiting headers)"},
		{resp: `{"asset":{"key":"assets/hello.txt"}}`, code: 200},
		{code: 404, err: ErrNotPartOfTheme.Error()},
	}

	for _, testcase := range testcases {
		m := new(mocks.HttpAdapter)
		client, _ := NewClient(&env.Env{ThemeID: "123"})
		client.http = m

		expectation := m.On("Get", APIPath+"themes/123/assets.json?asset%5Bkey%5D=filename.txt", NoHeaders)
		if testcase.resperr != "" {
			expectation.Return(nil, errors.New(testcase.resperr))
		} else if testcase.code != 0 {
			expectation.Return(jsonResponse(testcase.resp, testcase.code), nil)
		}

		asset, err := client.GetAsset("filename.txt")

		if testcase.err == "" {
			assert.Nil(t, err)
			assert.Equal(t, asset.Key, "assets/hello.txt")
		} else if assert.NotNil(t, err, testcase.err) {
			assert.Contains(t, err.Error(), testcase.err)
		}

		m.AssertExpectations(t)
	}
}

func TestThemeClient_UpdateAsset(t *testing.T) {
	testcases := []struct {
		resp, resperr, err string
		code               int
	}{
		{resp: `{"errors": "Not Found"}`, code: 200, err: "Not Found"},
		{resperr: "(Client.Timeout exceeded while awaiting headers)", err: "(Client.Timeout exceeded while awaiting headers)"},
		{resp: `{"asset":{"key":"assets/hello.txt"}}`, code: 200},
		{resp: "{}", code: 404, err: ErrNotPartOfTheme.Error()},
	}

	for _, testcase := range testcases {
		m := new(mocks.HttpAdapter)
		client, _ := NewClient(&env.Env{ThemeID: "123"})
		client.http = m

		expectation := m.On("Put", APIPath+"themes/123/assets.json", map[string]Asset{"asset": {Key: "filename.txt"}}, map[string]string{})
		if testcase.resperr != "" {
			expectation.Return(nil, errors.New(testcase.resperr))
		} else if testcase.code != 0 {
			expectation.Return(jsonResponse(testcase.resp, testcase.code), nil)
		}

		err := client.UpdateAsset(Asset{Key: "filename.txt"}, "")

		if testcase.err == "" {
			assert.Nil(t, err)
		} else if assert.NotNil(t, err, testcase.err) {
			assert.Contains(t, err.Error(), testcase.err)
		}

		m.AssertExpectations(t)
	}

	m := new(mocks.HttpAdapter)
	client, _ := NewClient(&env.Env{ThemeID: "123"})
	client.http = m
	asset := Asset{Key: "filename.txt"}

	count := 0
	m.On(
		"Put",
		mock.MatchedBy(func(path string) bool {
			if count == 0 {
				count++
				return true
			}
			return false
		}),
		map[string]Asset{"asset": asset},
		map[string]string{},
	).Return(&http.Response{
		Body:       &StringReadCloser{strings.NewReader(`{"errors":{"asset":["Cannot overwrite generated asset filename.txt"]}}`)},
		StatusCode: 422,
	}, nil)

	m.On(
		"Delete",
		APIPath+"themes/123/assets.json?asset%5Bkey%5D=filename.txt.liquid",
		NoHeaders,
	).Return(jsonResponse("{}", 200), nil)

	m.On(
		"Put",
		APIPath+"themes/123/assets.json",
		map[string]Asset{"asset": asset},
		map[string]string{},
	).Return(jsonResponse(`{"asset":{"key":"assets/hello.txt"}}`, 200), nil)

	assert.Nil(t, client.UpdateAsset(asset, ""))
	m.AssertExpectations(t)
}

func TestThemeClient_DeleteAsset(t *testing.T) {
	testcases := []struct {
		code               int
		resp, resperr, err string
	}{
		{resp: `{"errors": "server error"}`, err: "server error"},
		{resperr: "(Client.Timeout exceeded while awaiting headers)", err: "(Client.Timeout exceeded while awaiting headers)"},
		{code: 404, err: ErrNotPartOfTheme.Error()},
		{resp: `{}`, code: 200},
	}

	for _, testcase := range testcases {
		m := new(mocks.HttpAdapter)
		client, _ := NewClient(&env.Env{ThemeID: "123"})
		client.http = m

		expectation := m.On("Delete", APIPath+"themes/123/assets.json?asset%5Bkey%5D=filename.txt", NoHeaders)
		if testcase.resperr != "" {
			expectation.Return(nil, errors.New(testcase.resperr))
		} else {
			expectation.Return(jsonResponse(testcase.resp, testcase.code), nil)
		}

		err := client.DeleteAsset(Asset{Key: "filename.txt"})

		if testcase.err == "" {
			assert.Nil(t, err)
		} else if assert.NotNil(t, err, testcase.err) {
			assert.Contains(t, err.Error(), testcase.err)
		}

		m.AssertExpectations(t)
	}
}

func TestThemeClient_assetPath(t *testing.T) {
	testcases := []struct {
		query         map[string]string
		themeID, path string
	}{
		{themeID: "123", path: APIPath + "themes/123/assets.json?asset%5Bkey%5D=layout%2Ftheme.liquid", query: map[string]string{"asset[key]": "layout/theme.liquid"}},
		{path: "/admin/api/unstable/assets.json?asset%5Bkey%5D=layout%2Ftheme.liquid", query: map[string]string{"asset[key]": "layout/theme.liquid"}},
		{themeID: "123", path: APIPath + "themes/123/assets.json"},
		{path: "/admin/api/unstable/assets.json"},
	}

	for _, testcase := range testcases {
		client, _ := NewClient(&env.Env{ThemeID: testcase.themeID})
		path := client.assetPath(testcase.query)
		assert.Equal(t, testcase.path, path)
	}
}

func TestToMessages(t *testing.T) {
	testcases := []struct {
		input    map[string][]string
		expected []string
	}{
		{input: map[string][]string{"src": {"is empty"}}, expected: []string{"src is empty"}},
		{input: map[string][]string{}, expected: []string{}},
		{input: map[string][]string{"name": {"can't be blank"}}, expected: []string{"name can't be blank"}},
	}

	for _, testcase := range testcases {
		actual := toMessages(testcase.input)
		assert.Equal(t, testcase.expected, actual)
	}
}

func TestToSentence(t *testing.T) {
	testcases := []struct {
		input    []string
		expected string
	}{
		{input: []string{}, expected: ""},
		{input: []string{"src is empty"}, expected: "src is empty"},
		{input: []string{"src is empty", "name can't be blank"}, expected: "src is empty and name can't be blank"},
		{input: []string{"src is empty", "name can't be blank", "role is invalid"}, expected: "src is empty, name can't be blank, and role is invalid"},
	}

	for _, testcase := range testcases {
		actual := toSentence(testcase.input)
		assert.Equal(t, testcase.expected, actual)
	}
}

type StringReadCloser struct {
	*strings.Reader
}

func (s *StringReadCloser) Close() error {
	return nil
}

func jsonResponse(body string, code int) *http.Response {
	return &http.Response{
		Body:       &StringReadCloser{strings.NewReader(body)},
		StatusCode: code,
	}
}
