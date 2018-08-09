package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/Shopify/themekit/src/shopify"
)

func TestUpload(t *testing.T) {
	ctx, client, _, _, _ := createTestCtx()
	ctx.Args = []string{"templates/layout.liquid"}
	client.On("UpdateAsset", shopify.Asset{Key: "templates/layout.liquid"}).Return(nil)
	err := upload(ctx)
	assert.NotNil(t, err)

	ctx, client, _, _, _ = createTestCtx()
	ctx.Args = []string{"templates/layout.liquid"}
	ctx.Env.ReadOnly = true
	client.On("UpdateAsset", shopify.Asset{Key: "templates/layout.liquid"}).Return(nil)
	err = upload(ctx)
	if assert.NotNil(t, err) {
		assert.Contains(t, err.Error(), "environment is readonly")
	}

	ctx, client, _, stdOut, _ := createTestCtx()
	ctx.Args = []string{"assets/app.js"}
	ctx.Env.Directory = "_testdata/projectdir"
	ctx.Flags.Verbose = true
	client.On("UpdateAsset", shopify.Asset{Key: "assets/app.js"}).Return(nil)
	err = upload(ctx)
	assert.Nil(t, err)
	assert.Contains(t, stdOut.String(), "Updated assets/app.js")

	ctx, client, _, stdOut, _ = createTestCtx()
	ctx.Env.Directory = "_testdata/projectdir"
	ctx.Flags.Verbose = true
	client.On("UpdateAsset", mock.MatchedBy(func(a shopify.Asset) bool { return true })).Return(nil)
	err = upload(ctx)
	assert.Nil(t, err)
	assert.Contains(t, stdOut.String(), "Updated config/settings_data.json")
}
