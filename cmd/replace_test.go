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

func TestReplace(t *testing.T) {
	ctx, client, _, stdOut, _ := createTestCtx()
	ctx.Flags.Verbose = true
	ctx.Env.Directory = filepath.Join("_testdata", "projectdir")
	client.On("GetAllAssets").Return([]string{"assets/logo.png"}, nil)
	client.On("UpdateAsset", mock.MatchedBy(func(shopify.Asset) bool { return true })).Return(nil).Times(2)
	client.On("DeleteAsset", mock.MatchedBy(func(shopify.Asset) bool { return true })).Return(nil).Once()
	err := replace(ctx)
	assert.Nil(t, err)
	assert.Contains(t, stdOut.String(), "Updated config/settings_data.json")
	client.AssertExpectations(t)

	ctx, client, _, _, _ = createTestCtx()
	client.On("GetAllAssets").Return([]string{}, fmt.Errorf("server error"))
	err = replace(ctx)
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
