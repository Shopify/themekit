package kit

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestKitError(t *testing.T) {
	testErr := fmt.Errorf("testing error")
	err := kitError{err: testErr}
	assert.True(t, err.Fatal())
	assert.Equal(t, err.Error(), testErr.Error())
}

func TestThemeError(t *testing.T) {
	err := themeError{resp: ShopifyResponse{}}

	tests := map[int]bool{
		0:   true,
		200: false,
		404: true,
		500: true,
	}

	for code, check := range tests {
		err.resp.Code = code
		assert.Equal(t, err.Fatal(), check, fmt.Sprintf("code: %v, check: %v", code, check))
	}

	err.resp.Errors = requestError{Other: []string{"nope"}}
	err.resp.Code = 200
	assert.True(t, err.Fatal())
}

func TestAssetError(t *testing.T) {
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
		assert.Equal(t, err.Fatal(), check, fmt.Sprintf("code: %v, check: %v", code, check))
	}

	err.requestErr = requestError{Other: []string{"nope"}}
	err.resp.Code = 200
	assert.True(t, err.Fatal())
}

func TestListError(t *testing.T) {
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
		assert.Equal(t, err.Fatal(), check, fmt.Sprintf("code: %v, check: %v", code, check))
	}

	err.requestErr = requestError{Other: []string{"nope"}}
	err.resp.Code = 200
	assert.True(t, err.Fatal())
}
