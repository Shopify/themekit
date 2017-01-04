package kit

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type ErrorsTestSuite struct {
	suite.Suite
}

func (suite *ErrorsTestSuite) TestKitError() {
	testErr := fmt.Errorf("testing error")
	err := kitError{err: testErr}
	assert.True(suite.T(), err.Fatal())
	assert.Equal(suite.T(), err.Error(), testErr.Error())
}

func (suite *ErrorsTestSuite) TestThemeError() {
	err := themeError{resp: ShopifyResponse{}}

	tests := map[int]bool{
		0:   true,
		200: false,
		404: true,
		500: true,
	}

	for code, check := range tests {
		err.resp.Code = code
		assert.Equal(suite.T(), err.Fatal(), check, fmt.Sprintf("code: %v, check: %v", code, check))
	}

	err.resp.Errors = requestError{Other: []string{"nope"}}
	err.resp.Code = 200
	assert.True(suite.T(), err.Fatal())
}

func (suite *ErrorsTestSuite) TestAssetError() {
	err := assetError{resp: ShopifyResponse{}}

	tests := map[int]bool{
		0:   true,
		200: false,
		404: false,
		403: true,
		500: true,
	}

	for code, check := range tests {
		err.resp.Code = code
		assert.Equal(suite.T(), err.Fatal(), check, fmt.Sprintf("code: %v, check: %v", code, check))
	}

	err.requestErr = requestError{Other: []string{"nope"}}
	err.resp.Code = 200
	assert.True(suite.T(), err.Fatal())
}

func (suite *ErrorsTestSuite) TestListError() {
	err := listError{resp: ShopifyResponse{}}

	tests := map[int]bool{
		0:   true,
		200: false,
		304: false,
		404: true,
		403: true,
		500: true,
	}

	for code, check := range tests {
		err.resp.Code = code
		assert.Equal(suite.T(), err.Fatal(), check, fmt.Sprintf("code: %v, check: %v", code, check))
	}

	err.requestErr = requestError{Other: []string{"nope"}}
	err.resp.Code = 200
	assert.True(suite.T(), err.Fatal())
}

func (suite *ErrorsTestSuite) TestGeneralRequestError() {
}

func TestErrorsTestSuite(t *testing.T) {
	suite.Run(t, new(ErrorsTestSuite))
}
