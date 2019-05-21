package cmd

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/Shopify/themekit/src/shopify"
)

func TestDownload(t *testing.T) {
	allFilenames := []string{"assets/logo.png", "templates/customers/test.liquid", "config/test.liquid", "layout/test.liquid", "snippets/test.liquid", "templates/test.liquid", "locales/test.liquid", "sections/test.liquid"}

	ctx, client, _, _, stdErr := createTestCtx()
	client.On("GetAllAssets").Return(allFilenames, nil)
	client.On("GetAsset", mock.MatchedBy(func(string) bool { return true })).Return(shopify.Asset{}, nil).Times(len(allFilenames))
	err := download(ctx)
	assert.Nil(t, err)
	assert.Contains(t, stdErr.String(), "error writing assets/logo.png")

	ctx, client, _, _, _ = createTestCtx()
	client.On("GetAllAssets").Return(allFilenames, fmt.Errorf("server error"))
	err = download(ctx)
	if assert.NotNil(t, err) {
		assert.Contains(t, err.Error(), "server error")
	}

	ctx, client, _, _, stdErr = createTestCtx()
	client.On("GetAllAssets").Return(allFilenames, nil)
	client.On("GetAsset", mock.MatchedBy(func(string) bool { return true })).Return(shopify.Asset{}, fmt.Errorf("asset err"))
	assert.Nil(t, download(ctx))
	assert.Contains(t, stdErr.String(), "error downloading assets/logo.png")
}

func TestFilesToDownload(t *testing.T) {
	allFilenames := []string{"assets/logo.png", "templates/customers/test.liquid", "config/test.liquid", "layout/test.liquid", "snippets/test.liquid", "templates/test.liquid", "locales/test.liquid", "sections/test.liquid"}

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
		{args: []string{"assets/nope.png"}, ret: allFilenames, respErr: fmt.Errorf("server error"), err: "server error"},
	}

	for _, testcase := range testcases {
		ctx, client, _, _, _ := createTestCtx()
		ctx.Args = testcase.args
		client.On("GetAllAssets").Return(allFilenames, testcase.respErr)
		filenames, err := filesToDownload(ctx)
		assert.Equal(t, testcase.ret, filenames)
		if testcase.err == "" {
			assert.Nil(t, err)
		} else if assert.NotNil(t, err) {
			assert.Contains(t, err.Error(), testcase.err)
		}
	}
}
