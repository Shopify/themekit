package cmd

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/Shopify/themekit/src/file"
	"github.com/Shopify/themekit/src/shopify"
)

func TestUpload(t *testing.T) {
	ctx, client, _, _, _ := createTestCtx()
	ctx.Args = []string{"templates/layout.liquid"}
	ctx.Flags.NoDelete = true
	client.On("UpdateAsset", shopify.Asset{Key: "templates/layout.liquid"}).Return(nil)
	err := deploy(ctx)
	assert.NotNil(t, err)

	ctx, client, _, _, _ = createTestCtx()
	ctx.Args = []string{"templates/layout.liquid"}
	ctx.Flags.NoDelete = true
	ctx.Env.ReadOnly = true
	client.On("UpdateAsset", shopify.Asset{Key: "templates/layout.liquid"}).Return(nil)
	err = deploy(ctx)
	if assert.NotNil(t, err) {
		assert.Contains(t, err.Error(), "environment is readonly")
	}

	ctx, client, _, stdOut, _ := createTestCtx()
	ctx.Args = []string{"assets/app.js"}
	ctx.Flags.NoDelete = true
	ctx.Env.Directory = "_testdata/projectdir"
	ctx.Flags.Verbose = true
	client.On("UpdateAsset", shopify.Asset{Key: "assets/app.js"}).Return(nil)
	err = deploy(ctx)
	assert.Nil(t, err)
	assert.Contains(t, stdOut.String(), "Updated assets/app.js")

	ctx, client, _, stdOut, _ = createTestCtx()
	ctx.Env.Directory = "_testdata/projectdir"
	ctx.Flags.Verbose = true
	ctx.Flags.NoDelete = true
	client.On("UpdateAsset", mock.MatchedBy(func(a shopify.Asset) bool { return true })).Return(nil)
	err = deploy(ctx)
	assert.Nil(t, err)
	assert.Contains(t, stdOut.String(), "Updated config/settings_data.json")
}

func TestReplace(t *testing.T) {
	ctx, client, _, stdOut, _ := createTestCtx()
	ctx.Flags.Verbose = true
	ctx.Env.Directory = filepath.Join("_testdata", "projectdir")
	client.On("GetAllAssets").Return([]string{"assets/logo.png"}, nil)
	client.On("UpdateAsset", mock.MatchedBy(func(shopify.Asset) bool { return true })).Return(nil).Times(2)
	client.On("DeleteAsset", mock.MatchedBy(func(shopify.Asset) bool { return true })).Return(nil).Once()
	err := deploy(ctx)
	assert.Nil(t, err)
	assert.Contains(t, stdOut.String(), "Updated config/settings_data.json")
	client.AssertExpectations(t)

	ctx, client, _, _, _ = createTestCtx()
	client.On("GetAllAssets").Return([]string{}, fmt.Errorf("server error"))
	err = deploy(ctx)
	if assert.NotNil(t, err) {
		assert.Contains(t, err.Error(), "server error")
	}
}

func TestGenerateActions(t *testing.T) {
	ctx, client, _, _, _ := createTestCtx()
	ctx.Env.Directory = filepath.Join("_testdata", "projectdir")
	client.On("GetAllAssets").Return([]string{"assets/logo.png"}, nil)
	actions, err := generateActions(ctx)
	assert.Nil(t, err)
	assert.Equal(t, actions["assets/logo.png"], file.Remove)
	assert.Equal(t, actions["config/settings_data.json"], file.Update)
	assert.Equal(t, actions["assets/app.js"], file.Update)
	assert.Equal(t, len(actions), 3)
	_, found := actions["assets/.gitkeep"]
	assert.False(t, found)

	ctx, client, _, _, _ = createTestCtx()
	ctx.Env.Directory = filepath.Join("_testdata", "projectdir")
	client.On("GetAllAssets").Return([]string{}, fmt.Errorf("server error"))
	_, err = generateActions(ctx)
	if assert.NotNil(t, err) {
		assert.Contains(t, err.Error(), "server error")
	}

	ctx, client, _, _, _ = createTestCtx()
	ctx.Env.Directory = "not there"
	client.On("GetAllAssets").Return([]string{}, nil)
	_, err = generateActions(ctx)
	assert.NotNil(t, err)
}
