package shopify

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"os"
	"text/template"
)

// reqErr is the expected response structure of errors from asset operations.
type reqErr struct {
	Errors string `json:"errors"`
}

// Err will return an error if the struct contains any errors, otherwise it will return nil
func (err reqErr) Err() error {
	if len(err.Errors) > 0 {
		return errors.New(err.Errors)
	}
	return nil
}

// RespUnmarshalError is an error struct to allow us to present response issues
// in a more intelligent way, and in a way that will make analyzing the errors
// easier.
type RespUnmarshalError struct {
	Resp       *http.Response
	Problem    string
	Suggestion string
	ReadErr    error
	TmpFile    *os.File
}

var respUnmarshalErrorTemplate = template.Must(template.New("respUnmarshalError").Parse(
	`{{ .Problem }}
{{ .Suggestion }}
Http Response Status: {{ .Resp.StatusCode }}
Request ID: {{ .RequestID }}{{ if .ReadErr }}
Error: {{ .ReadErr }}{{ end }}{{ with .TmpFile }}
ResponseBody: {{ .Name }}{{ end }}`))

// RequestID is a helper for the template
func (err RespUnmarshalError) RequestID() string {
	return err.Resp.Header.Get("X-Request-Id")
}

// Error satisfies the Error interface
func (err RespUnmarshalError) Error() string {
	var tpl bytes.Buffer
	respUnmarshalErrorTemplate.Execute(&tpl, err)
	return tpl.String()
}

// unmarshalResponse is the bottle neck of receiving an http response and unmarshalling
// into assets or theme data. If this data cannot be successfully unmarshalled, then
// it will fallback to try to unmarshal expected error responses in the JSON format.
// If this finally does not work, then this will return an error of what went wrong
// and dump the response body into a temporary file for later examination.
func unmarshalResponse(resp *http.Response, data interface{}) error {
	reqBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return RespUnmarshalError{
			Resp:       resp,
			Problem:    "could not read response body",
			Suggestion: "This may mean that the request was not able to finish successfully",
			ReadErr:    err,
		}
	}
	defer resp.Body.Close()

	var re reqErr
	mainErr := json.Unmarshal(reqBody, data) // check if we can unmarshal into the expected returned data
	basicErr := json.Unmarshal(reqBody, &re) // if no returned data, check if we can get errors from the body
	if mainErr != nil && basicErr != nil {
		tmpFile, err := ioutil.TempFile(os.TempDir(), "themekit-response-*.txt")
		if err == nil {
			defer tmpFile.Close()
			tmpFile.Write([]byte(reqBody))
		}
		return RespUnmarshalError{
			Resp:       resp,
			Problem:    "could not unmarshal JSON from response body",
			Suggestion: "This usually means Theme Kit received an HTML error page.",
			TmpFile:    tmpFile,
		}
	}

	return re.Err()
}
