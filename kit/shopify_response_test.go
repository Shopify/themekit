package kit

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func mockShopifyResp(body string) (*ShopifyResponse, Error) {
	req := newShopifyRequest(&Configuration{}, themeRequest, Create, "")
	return newShopifyResponse(req, &http.Response{
		Status:     "200 OK",
		StatusCode: 200,
		Request:    &http.Request{URL: &url.URL{}},
		Body:       ioutil.NopCloser(strings.NewReader(body)),
	}, nil)
}

func TestNewShopifyResponse(t *testing.T) {
	errorMessage := "something went wrong"
	badErr := fmt.Errorf(errorMessage)
	req := newShopifyRequest(&Configuration{}, themeRequest, Create, "")
	resp, err := newShopifyResponse(req, nil, badErr)
	assert.NotNil(t, resp)
	assert.NotNil(t, err)

	mock := &http.Response{
		Request: &http.Request{URL: &url.URL{}},
		Body:    ioutil.NopCloser(strings.NewReader("")),
	}
	mock.Body.Close()
	req = newShopifyRequest(&Configuration{}, themeRequest, Create, "")
	resp, err = newShopifyResponse(req, mock, nil)
	assert.NotNil(t, resp)
	assert.NotNil(t, err)

	resp, err = mockShopifyResp(`{"errors": "this is api error"}`)
	assert.NotNil(t, resp)
	if assert.NotNil(t, err) {
		assert.Equal(t, "this is api error", resp.Errors.Error())
	}

	resp, err = mockShopifyResp(`{"theme":{"name":"timberland"}}`)
	assert.NotNil(t, resp)
	assert.Nil(t, err)
	assert.Equal(t, "timberland", resp.Theme.Name)

	resp, err = mockShopifyResp(`{"errors":{"src":["is empty"]}}`)
	assert.NotNil(t, resp)
	if assert.NotNil(t, err) {
		assert.Equal(t, "is empty", resp.Errors.Error())
	}

	resp, err = mockShopifyResp(`{"asset":{"key": "assets/hello.txt"}}`)
	assert.NotNil(t, resp)
	assert.Nil(t, err)
	assert.Equal(t, "assets/hello.txt", resp.Asset.Key)

	resp, err = mockShopifyResp(`{"errors":{"asset":["this is asset error"]}}`)
	assert.NotNil(t, resp)
	if assert.NotNil(t, err) {
		assert.Equal(t, "this is asset error", resp.Errors.Error())
	}

	resp, err = mockShopifyResp(`{"assets":[{"key":"assets/hello.txt"},{"key":"assets/goodbye.txt"}]}`)
	assert.NotNil(t, resp)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(resp.Assets))
	assert.Equal(t, "assets/goodbye.txt", resp.Assets[1].Key)
}

func TestShopifyResponse_Successful(t *testing.T) {
	resp := ShopifyResponse{Code: 200}
	assert.Equal(t, true, resp.Successful())
	resp = ShopifyResponse{Code: 500}
	assert.Equal(t, false, resp.Successful())
	resp = ShopifyResponse{Code: 200, Errors: requestError{Other: []string{"nope"}}}
	assert.Equal(t, false, resp.Successful())
}

func TestShopifyResponse_Error(t *testing.T) {
	resp := ShopifyResponse{Code: 200}
	assert.Nil(t, resp.Error())
	resp = ShopifyResponse{Code: 500, Type: themeRequest}
	assert.IsType(t, themeError{}, resp.Error())
	resp = ShopifyResponse{Code: 500, Type: assetRequest}
	assert.IsType(t, assetError{}, resp.Error())
	resp = ShopifyResponse{Code: 500, Type: listRequest}
	assert.IsType(t, listError{}, resp.Error())
	resp = ShopifyResponse{Code: 500, Type: 20}
	assert.IsType(t, kitError{}, resp.Error())
}
