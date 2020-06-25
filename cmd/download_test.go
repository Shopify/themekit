package cmd

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/Shopify/themekit/src/shopify"
)

func TestDownload(t *testing.T) {
	allAssets := []shopify.Asset{
		{Key: "assets/logo.png"},
		{Key: "templates/customers/test.liquid"},
		{Key: "config/test.liquid"},
		{Key: "layout/test.liquid"},
		{Key: "snippets/test.liquid"},
		{Key: "templates/test.liquid"},
		{Key: "locales/test.liquid"},
		{Key: "sections/test.liquid"},
	}

	ctx, client, _, _, stdErr := createTestCtx()
	client.On("GetAllAssets").Return(allAssets, nil)
	client.On("GetAsset", mock.MatchedBy(func(string) bool { return true })).Return(shopify.Asset{}, nil).Times(len(allAssets))
	err := download(ctx)
	assert.Nil(t, err)
	assert.Contains(t, stdErr.String(), "error writing assets/logo.png")

	ctx, client, _, _, _ = createTestCtx()
	client.On("GetAllAssets").Return(allAssets, fmt.Errorf("server error"))
	err = download(ctx)
	if assert.NotNil(t, err) {
		assert.Contains(t, err.Error(), "server error")
	}

	ctx, client, _, _, stdErr = createTestCtx()
	client.On("GetAllAssets").Return(allAssets, nil)
	client.On("GetAsset", mock.MatchedBy(func(string) bool { return true })).Return(shopify.Asset{}, fmt.Errorf("asset err"))
	assert.Nil(t, download(ctx))
	assert.Contains(t, stdErr.String(), "error downloading assets/logo.png")
}

func TestFilesToDownload(t *testing.T) {
	allAssets := []shopify.Asset{{Key: "assets/logo.png"}, {Key: "templates/customers/test.liquid"}, {Key: "config/test.liquid"}, {Key: "layout/test.liquid"}, {Key: "snippets/test.liquid"}, {Key: "templates/test.liquid"}, {Key: "locales/test.liquid"}, {Key: "sections/test.liquid"}}
	allFilenames := []string{}
	for _, asset := range allAssets {
		allFilenames = append(allFilenames, asset.Key)
	}

	testcases := []struct {
		err       string
		respErr   error
		args, ret []string
	}{
		{ret: allFilenames},
		{args: []string{"assets/logo.png"}, ret: []string{"assets/logo.png"}},
		{args: []string{"assets/*"}, ret: []string{"assets/logo.png"}},
		{args: []string{"templates"}, ret: []string{"templates/test.liquid"}},
		{args: []string{"assets/nope.png"}, ret: []string{}, err: "No file paths matched the inputted arguments"},
		{args: []string{"assets/nope.png"}, ret: []string{}, respErr: fmt.Errorf("server error"), err: "server error"},
	}

	for i, testcase := range testcases {
		ctx, client, _, _, _ := createTestCtx()
		ctx.Args = testcase.args
		client.On("GetAllAssets").Return(allAssets, testcase.respErr)
		filenames, err := filesToDownload(ctx)
		assert.Equal(t, testcase.ret, filenames, fmt.Sprintf("Failed to compare filenames in test case %d", i))
		if testcase.err == "" {
			assert.Nil(t, err)
		} else if assert.NotNil(t, err) {
			assert.Contains(t, err.Error(), testcase.err)
		}
	}
}
