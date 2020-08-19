package cmd

import (
	"bytes"
	"fmt"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/Shopify/themekit/src/colors"
	"github.com/Shopify/themekit/src/file"
	"github.com/Shopify/themekit/src/shopify"
)

func TestUploadSingleFile(t *testing.T) {
	ctx, client, _, _, _ := createTestCtx()
	ctx.Args = []string{"templates/layout.liquid"}
	ctx.Flags.NoDelete = true
	client.On("GetAllAssets").Return([]shopify.Asset{{Key: "templates/layout.liquid"}}, nil)
	client.On("UpdateAsset", shopify.Asset{Key: "templates/layout.liquid"}, "").Return(nil)
	err := deploy(ctx)
	assert.NotNil(t, err)
}

func TestUploadWithReadOnlyOption(t *testing.T) {
	ctx, client, _, _, _ := createTestCtx()
	ctx.Args = []string{"templates/layout.liquid"}
	ctx.Flags.NoDelete = true
	ctx.Env.ReadOnly = true
	client.On("UpdateAsset", shopify.Asset{Key: "templates/layout.liquid"}, "").Return(nil)
	err := deploy(ctx)
	if assert.NotNil(t, err) {
		assert.Contains(t, err.Error(), "environment is readonly")
	}
}

func TestUploadWithVerboseOptions(t *testing.T) {
	ctx, client, _, stdOut, _ := createTestCtx()
	ctx.Args = []string{"assets/app.js"}
	ctx.Flags.NoDelete = true
	ctx.Env.Directory = "_testdata/projectdir"
	ctx.Flags.Verbose = true
	client.On("GetAllAssets").Return([]shopify.Asset{{Key: "assets/app.js"}}, nil)
	// This checksum corresponds to a zero-byte file
	client.On("UpdateAsset", shopify.Asset{Key: "assets/app.js", Checksum: "d41d8cd98f00b204e9800998ecf8427e"}, "").Return(nil)
	err := deploy(ctx)
	assert.Nil(t, err)
	assert.Contains(t, stdOut.String(), "Updated assets/app.js")
}

func TestUploadAllFiles(t *testing.T) {
	ctx, client, _, stdOut, _ := createTestCtx()
	ctx.Env.Directory = "_testdata/projectdir"
	ctx.Flags.Verbose = true
	ctx.Flags.NoDelete = true
	client.On("GetAllAssets").Return([]shopify.Asset{{Key: "config/settings_data.json"}}, nil)
	client.On("UpdateAsset", mock.MatchedBy(func(a shopify.Asset) bool { return true }), "").Return(nil)
	err := deploy(ctx)
	assert.Nil(t, err)
	assert.Contains(t, stdOut.String(), "Updated config/settings_data.json")
}

func TestUploadForSkipFileWhenChecksumsMatch(t *testing.T) {
	ctx, client, _, stdOut, _ := createTestCtx()
	ctx.Env.Directory = "_testdata/projectdir"
	ctx.Flags.Verbose = true
	ctx.Flags.NoDelete = true
	client.On("GetAllAssets").Return([]shopify.Asset{{Key: "config/settings_data.json", Checksum: "d41d8cd98f00b204e9800998ecf8427e"}}, nil)
	// the _testdirectory contains two assets. We expect one to be uploaded, one to be skipped.
	client.On("UpdateAsset", shopify.Asset{Key: "assets/app.js", Checksum: "d41d8cd98f00b204e9800998ecf8427e"}, "").Return(nil)
	err := deploy(ctx)
	assert.Nil(t, err)
	assert.Contains(t, stdOut.String(), "Skipped config/settings_data.json")
	assert.Contains(t, stdOut.String(), "Updated assets/app.js")
}

func TestUploadForDoNotSkipWhenChecksumsDiffer(t *testing.T) {
	ctx, client, _, stdOut, _ := createTestCtx()
	ctx.Env.Directory = "_testdata/projectdir"
	ctx.Flags.Verbose = true
	ctx.Flags.NoDelete = true
	client.On("GetAllAssets").Return([]shopify.Asset{{Key: "config/settings_data.json", Checksum: "abc123"}}, nil)
	client.On("UpdateAsset", mock.MatchedBy(func(a shopify.Asset) bool { return true }), "").Return(nil)
	err := deploy(ctx)
	assert.Nil(t, err)
	assert.Contains(t, stdOut.String(), "Updated config/settings_data.json")
}

func TestReplace(t *testing.T) {
	ctx, client, _, stdOut, _ := createTestCtx()
	ctx.Flags.Verbose = true
	ctx.Env.Directory = filepath.Join("_testdata", "projectdir")
	client.On("GetAllAssets").Return([]shopify.Asset{{Key: "assets/logo.png"}}, nil)
	client.On("UpdateAsset", mock.MatchedBy(func(shopify.Asset) bool { return true }), "").Return(nil).Times(2)
	client.On("DeleteAsset", mock.MatchedBy(func(shopify.Asset) bool { return true })).Return(nil).Once()
	err := deploy(ctx)
	assert.Nil(t, err)
	assert.Contains(t, stdOut.String(), "Updated config/settings_data.json")
	client.AssertExpectations(t)

	ctx, client, _, _, _ = createTestCtx()
	client.On("GetAllAssets").Return([]shopify.Asset{}, fmt.Errorf("server error"))
	err = deploy(ctx)
	if assert.NotNil(t, err) {
		assert.Contains(t, err.Error(), "server error")
	}
}

func TestGenerateActions(t *testing.T) {
	ctx, client, _, _, _ := createTestCtx()
	ctx.Env.Directory = filepath.Join("_testdata", "projectdir")
	client.On("GetAllAssets").Return([]shopify.Asset{{Key: "assets/logo.png"}}, nil)
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
	client.On("GetAllAssets").Return([]shopify.Asset{}, fmt.Errorf("server error"))
	_, err = generateActions(ctx)
	if assert.NotNil(t, err) {
		assert.Contains(t, err.Error(), "server error")
	}

	ctx, client, _, _, _ = createTestCtx()
	ctx.Env.Directory = "not there"
	client.On("GetAllAssets").Return([]shopify.Asset{}, nil)
	_, err = generateActions(ctx)
	assert.NotNil(t, err)

	ctx, client, _, _, _ = createTestCtx()
	ctx.Env.Name = "development"
	ctx.Env.Directory = filepath.Join("_testdata", "badprojectdir")
	client.On("GetAllAssets").Return([]shopify.Asset{}, nil)
	actions, err = generateActions(ctx)
	assert.NotNil(t, err)
	var tpl bytes.Buffer
	compiledFilenameWarning.Execute(&tpl, struct {
		EnvName   string
		FileNames []string
	}{EnvName: colors.Yellow("development"), FileNames: []string{colors.Yellow("assets/app.js") + colors.Blue(" conflicts with ") + colors.Yellow("assets/app.js.liquid")}})
	assert.Equal(t, tpl.String(), err.Error())
}

func TestCompileAssetFilenames(t *testing.T) {
	input := []shopify.Asset{
		{Key: "assets/app.js"},
		{Key: "assets/app.scss"},
		{Key: "assets/foo.js.liquid"},
		{Key: "assets/app.js.liquid"},
		{Key: "assets/foo.js"},
	}

	expected := []string{
		colors.Yellow("assets/app.js") + colors.Blue(" conflicts with ") + colors.Yellow("assets/app.js.liquid"),
		colors.Yellow("assets/foo.js") + colors.Blue(" conflicts with ") + colors.Yellow("assets/foo.js.liquid"),
	}
	assert.Equal(t, expected, compileAssetFilenames(input))
}

func TestCompiledAssetWarning(t *testing.T) {
	filenames := []string{
		colors.Yellow("assets/app.js") + colors.Blue(" conflicts with ") + colors.Yellow("assets/app.js.liquid"),
	}

	var tpl bytes.Buffer
	compiledFilenameWarning.Execute(&tpl, struct {
		EnvName   string
		FileNames []string
	}{EnvName: colors.Yellow("development"), FileNames: filenames})

	assert.Equal(t, tpl.String(), compiledAssetWarning("development", filenames).Error())
}
