package kit

import (
	"fmt"
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type ShopifyResponseTestSuite struct {
	suite.Suite
}

func (suite *ShopifyResponseTestSuite) TestRequestError() {
	errorMessage := "something went wrong"
	badErr := fmt.Errorf(errorMessage)
	req, _ := newShopifyRequest(newTestConfig(), themeRequest, Create, "")
	resp, err := newShopifyResponse(req, nil, badErr)
	assert.NotNil(suite.T(), resp)
	assert.NotNil(suite.T(), err)
}

func (suite *ShopifyResponseTestSuite) TestNoBody() {
	mock := &http.Response{
		Request: &http.Request{URL: &url.URL{}},
		Body:    fileFixture("responses/general_error"),
	}
	mock.Body.Close()
	req, _ := newShopifyRequest(newTestConfig(), themeRequest, Create, "")
	resp, err := newShopifyResponse(req, mock, nil)
	assert.NotNil(suite.T(), resp)
	assert.NotNil(suite.T(), err)
}

func (suite *ShopifyResponseTestSuite) TestErrorResponse() {
	resp, err := suite.shopifyResp("responses/general_error")
	assert.NotNil(suite.T(), resp)
	if assert.NotNil(suite.T(), err) {
		assert.Equal(suite.T(), "[API] Invalid API key or access token (unrecognized login or wrong password)", resp.Errors.Error())
	}
}

func (suite *ShopifyResponseTestSuite) TestThemeResponse() {
	resp, err := suite.shopifyResp("responses/theme")
	assert.NotNil(suite.T(), resp)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), "timberland", resp.Theme.Name)
}

func (suite *ShopifyResponseTestSuite) TestThemeErrorResponse() {
	resp, err := suite.shopifyResp("responses/theme_error")
	assert.NotNil(suite.T(), resp)
	if assert.NotNil(suite.T(), err) {
		assert.Equal(suite.T(), "is empty", resp.Errors.Error())
	}
}

func (suite *ShopifyResponseTestSuite) TestAssetResponse() {
	resp, err := suite.shopifyResp("responses/asset")
	assert.NotNil(suite.T(), resp)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), "assets/hello.txt", resp.Asset.Key)
}

func (suite *ShopifyResponseTestSuite) TestAssetErrorResponse() {
	resp, err := suite.shopifyResp("responses/asset_error")
	assert.NotNil(suite.T(), resp)
	if assert.NotNil(suite.T(), err) {
		assert.Equal(suite.T(), "Liquid syntax error (line 10): 'comment' tag was never closed", resp.Errors.Error())
	}
}

func (suite *ShopifyResponseTestSuite) TestListResponse() {
	resp, err := suite.shopifyResp("responses/assets")
	assert.NotNil(suite.T(), resp)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), 2, len(resp.Assets))
	assert.Equal(suite.T(), "assets/goodbye.txt", resp.Assets[1].Key)
}

func (suite *ShopifyResponseTestSuite) TestSuccessful() {
	resp := ShopifyResponse{Code: 200}
	assert.Equal(suite.T(), true, resp.Successful())
	resp = ShopifyResponse{Code: 500}
	assert.Equal(suite.T(), false, resp.Successful())
	resp = ShopifyResponse{Code: 200, Errors: requestError{Other: []string{"nope"}}}
	assert.Equal(suite.T(), false, resp.Successful())
}

func (suite *ShopifyResponseTestSuite) TestError() {
	resp := ShopifyResponse{Code: 200}
	assert.Nil(suite.T(), resp.Error())
	resp = ShopifyResponse{Code: 500, Type: themeRequest}
	assert.IsType(suite.T(), themeError{}, resp.Error())
	resp = ShopifyResponse{Code: 500, Type: assetRequest}
	assert.IsType(suite.T(), assetError{}, resp.Error())
	resp = ShopifyResponse{Code: 500, Type: listRequest}
	assert.IsType(suite.T(), listError{}, resp.Error())
	resp = ShopifyResponse{Code: 500, Type: 20}
	assert.IsType(suite.T(), kitError{}, resp.Error())
}

func (suite *ShopifyResponseTestSuite) shopifyResp(path string) (*ShopifyResponse, Error) {
	req, _ := newShopifyRequest(newTestConfig(), themeRequest, Create, "")
	return newShopifyResponse(req, suite.respFixture(path), nil)
}

func (suite *ShopifyResponseTestSuite) respFixture(path string) *http.Response {
	return &http.Response{
		Status:     "200 OK",
		StatusCode: 200,
		Request:    &http.Request{URL: &url.URL{}},
		Body:       fileFixture(path),
	}
}

func TestShopifyResponseTestSuite(t *testing.T) {
	suite.Run(t, new(ShopifyResponseTestSuite))
}
