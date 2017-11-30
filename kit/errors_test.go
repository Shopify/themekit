package kit

import (
	"fmt"
	"strings"
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
	assert.Equal(t, "none", err.requestErr.String())

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

	err.resp.Code = 403
	err.resp.EventType = Remove
	err.generateHints()
	assert.Equal(t, "This file is critical and removing it would cause your theme to become non-functional.", err.requestErr.Other[len(err.requestErr.Other)-1])

	err.resp.Code = 404
	err.resp.EventType = Update
	err.generateHints()
	assert.Equal(t, "This file is not part of your theme.", err.requestErr.Other[len(err.requestErr.Other)-1])

	err.resp.Code = 409
	err.resp.EventType = Update
	err.generateHints()
	assert.Equal(t, `There have been changes to this file made remotely.

You can solve this by running 'theme download' to get the most recent copy of this file.
Running 'theme download' will overwrite any changes you have made so make sure to make
a backup.

If you are certain that you want to overwrite any changes then use the --force flag`, err.requestErr.Other[len(err.requestErr.Other)-1])
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
	assert.True(t, strings.Contains(err.Error(), "Assets Perform"))
}
