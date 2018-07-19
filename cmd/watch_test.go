package cmd

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Shopify/themekit/src/file"
	"github.com/Shopify/themekit/src/shopify"
)

func TestWatch(t *testing.T) {
	ctx, _, _, _, _ := createTestCtx()
	ctx.Env.ReadOnly = true
	err := watch(ctx, make(chan file.Event), make(chan os.Signal))
	if assert.NotNil(t, err) {
		assert.Contains(t, err.Error(), "environment is reaonly")
	}

	eventChan := make(chan file.Event, 1)
	ctx, _, _, stdOut, _ := createTestCtx()
	ctx.Flags.ConfigPath = "config.yml"
	eventChan <- file.Event{Path: ctx.Flags.ConfigPath}
	err = watch(ctx, eventChan, make(chan os.Signal))
	if assert.NotNil(t, err) {
		assert.Contains(t, err.Error(), "reload")
	}
	assert.Contains(t, stdOut.String(), "Watching for file changes on host")
	assert.Contains(t, stdOut.String(), "Reloading config changes")

	signalChan := make(chan os.Signal)
	eventChan = make(chan file.Event)
	ctx, _, _, stdOut, stdErr := createTestCtx()
	ctx.Flags.ConfigPath = "config.yml"
	go func() {
		eventChan <- file.Event{Op: file.Update, Path: "assets/app.js"}
		signalChan <- os.Interrupt
	}()
	err = watch(ctx, eventChan, signalChan)
	assert.Nil(t, err)
	assert.Contains(t, stdOut.String(), "Watching for file changes on host")
	assert.Contains(t, stdOut.String(), "processing assets/app.js")
	assert.Contains(t, stdErr.String(), "error loading assets/app.js: readAsset: open assets/app.js")

	signalChan = make(chan os.Signal)
	eventChan = make(chan file.Event)
	ctx, client, _, stdOut, stdErr := createTestCtx()
	client.On("UpdateAsset", shopify.Asset{Key: "assets/app.js"}).Return(nil)
	ctx.Flags.ConfigPath = "config.yml"
	ctx.Env.Directory = "_testdata/projectdir"
	go func() {
		eventChan <- file.Event{Op: file.Update, Path: "assets/app.js"}
		signalChan <- os.Interrupt
	}()
	err = watch(ctx, eventChan, signalChan)
	assert.Nil(t, err)
	assert.Contains(t, stdOut.String(), "Watching for file changes on host")
	assert.Contains(t, stdOut.String(), "processing assets/app.js")
	assert.Contains(t, stdOut.String(), "Updated assets/app.js")

	signalChan = make(chan os.Signal)
	eventChan = make(chan file.Event)
	ctx, client, _, stdOut, stdErr = createTestCtx()
	client.On("DeleteAsset", shopify.Asset{Key: "assets/app.js"}).Return(nil)
	ctx.Flags.ConfigPath = "config.yml"
	ctx.Env.Directory = "_testdata/projectdir"
	go func() {
		eventChan <- file.Event{Op: file.Remove, Path: "assets/app.js"}
		signalChan <- os.Interrupt
	}()
	err = watch(ctx, eventChan, signalChan)
	assert.Nil(t, err)
	assert.Contains(t, stdOut.String(), "Watching for file changes on host")
	assert.Contains(t, stdOut.String(), "processing assets/app.js")
	assert.Contains(t, stdOut.String(), "Deleted assets/app.js")
}
