package shopify

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRespUnmarshalError(t *testing.T) {
	resp := &http.Response{
		StatusCode: 442,
		Header:     http.Header{"X-Request-Id": []string{"abc-123-456"}},
	}

	err := RespUnmarshalError{
		Resp:       resp,
		Problem:    "this is your problem",
		Suggestion: "test your things",
	}

	assert.Equal(t, err.Error(), `this is your problem
test your things
Http Response Status: 442
Request ID: abc-123-456`)

	err = RespUnmarshalError{
		Resp:       resp,
		Problem:    "this is your problem",
		Suggestion: "test your things",
		ReadErr:    fmt.Errorf("Bad READ"),
	}

	assert.Equal(t, err.Error(), `this is your problem
test your things
Http Response Status: 442
Request ID: abc-123-456
Error: Bad READ`)

	err = RespUnmarshalError{
		Resp:       resp,
		Problem:    "this is your problem",
		Suggestion: "test your things",
	}
	err.TmpFile, _ = ioutil.TempFile(os.TempDir(), "themekit-response-*.txt")
	err.TmpFile.Write([]byte("body"))
	err.TmpFile.Close()
	defer os.Remove(err.TmpFile.Name())

	assert.Equal(t, err.Error(), fmt.Sprintf(`this is your problem
test your things
Http Response Status: 442
Request ID: abc-123-456
ResponseBody: %v`, err.TmpFile.Name()))
}

func TestUnmarshalResponse(t *testing.T) {
	testcases := []struct {
		input, err    string
		code          int
		out, expected themeResponse
	}{
		{input: `{"errors":{"name":["can't be blank"]}}`, code: 200, expected: themeResponse{Errors: map[string][]string{"name": {"can't be blank"}}}},
		{input: `{"errors": "Not Found"}`, code: 404, err: "Not Found"},
		{input: `{"theme":{"id": 123456}}`, code: 200, expected: themeResponse{Theme: Theme{ID: int64(123456)}}},
		{input: `{"theme":{"id": 123456}}`, code: 200, expected: themeResponse{Theme: Theme{ID: int64(123456)}}},
		{input: `<html><body>BAD ERROR</body></html>`, code: 500, err: "could not unmarshal JSON from response body"},
	}

	for _, testcase := range testcases {
		err := unmarshalResponse(jsonResponse(testcase.input, testcase.code), &testcase.out)
		assert.Equal(t, testcase.expected, testcase.out)
		if testcase.err == "" {
			assert.Nil(t, err)
		} else if assert.NotNil(t, err) {
			assert.Contains(t, err.Error(), testcase.err)
		}
	}

	out := assetsResponse{}
	err := unmarshalResponse(jsonResponse(`{"errors":"oh no"}`, 200), &out)
	assert.NotNil(t, err)
	assert.Equal(t, err.Error(), "oh no")
}
