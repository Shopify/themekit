package cmd

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/Shopify/themekit/src/cmdutil"
	"github.com/Shopify/themekit/src/file"
	"github.com/Shopify/themekit/src/shopify"
)

type testAdapter struct{ mock.Mock }

func (adapter *testAdapter) notify(ctx *cmdutil.Ctx, path string) {
	adapter.Called(ctx, path)
}

func TestWatch(t *testing.T) {
	ctx, _, _, _, _ := createTestCtx()
	ctx.Env.ReadOnly = true
	err := watch(ctx, make(chan file.Event), make(chan os.Signal), nil)
	if assert.NotNil(t, err) {
		assert.Contains(t, err.Error(), "environment is reaonly")
	}

	eventChan := make(chan file.Event, 1)
	ctx, _, _, stdOut, _ := createTestCtx()
	ctx.Flags.ConfigPath = "config.yml"
	eventChan <- file.Event{Path: ctx.Flags.ConfigPath}
	err = watch(ctx, eventChan, make(chan os.Signal), nil)
	if assert.NotNil(t, err) {
		assert.Contains(t, err.Error(), "reload")
	}
	assert.Contains(t, stdOut.String(), "Watching for file changes")
	assert.Contains(t, stdOut.String(), "Reloading config changes")

	signalChan := make(chan os.Signal)
	eventChan = make(chan file.Event)
	ctx, _, _, stdOut, stdErr := createTestCtx()
	ctx.Flags.ConfigPath = "config.yml"
	go func() {
		eventChan <- file.Event{Op: file.Update, Path: "assets/app.js"}
		signalChan <- os.Interrupt
	}()
	notifier := new(testAdapter)
	notifier.On("notify", ctx, "assets/app.js")
	err = watch(ctx, eventChan, signalChan, notifier)
	assert.Nil(t, err)
	assert.Contains(t, stdOut.String(), "Watching for file changes")
	assert.Contains(t, stdOut.String(), "processing assets/app.js")
	assert.Contains(t, stdErr.String(), "error loading assets/app.js: readAsset: ")
	notifier.AssertExpectations(t)

	signalChan = make(chan os.Signal)
	eventChan = make(chan file.Event)
	ctx, client, _, stdOut, stdErr := createTestCtx()
	client.On("UpdateAsset", shopify.Asset{Key: "assets/app.js", Checksum: "d41d8cd98f00b204e9800998ecf8427e"}, "").Return(nil)
	ctx.Flags.ConfigPath = "config.yml"
	ctx.Env.Directory = "_testdata/projectdir"
	go func() {
		eventChan <- file.Event{Op: file.Update, Path: "assets/app.js"}
		signalChan <- os.Interrupt
	}()
	notifier = new(testAdapter)
	notifier.On("notify", ctx, "assets/app.js")
	err = watch(ctx, eventChan, signalChan, notifier)
	assert.Nil(t, err)
	assert.Contains(t, stdOut.String(), "Watching for file changes")
	assert.Contains(t, stdOut.String(), "processing assets/app.js")
	assert.Contains(t, stdOut.String(), "Updated assets/app.js")
	notifier.AssertExpectations(t)

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
	notifier = new(testAdapter)
	notifier.On("notify", ctx, "assets/app.js")
	err = watch(ctx, eventChan, signalChan, notifier)
	assert.Nil(t, err)
	assert.Contains(t, stdOut.String(), "Watching for file changes")
	assert.Contains(t, stdOut.String(), "processing assets/app.js")
	assert.Contains(t, stdOut.String(), "Deleted assets/app.js")
	notifier.AssertExpectations(t)

	signalChan = make(chan os.Signal)
	eventChan = make(chan file.Event)
	ctx, client, _, stdOut, stdErr = createTestCtx()
	client.On("UpdateAsset", shopify.Asset{Key: "assets/app.js", Checksum: "d41d8cd98f00b204e9800998ecf8427e"}, "").Return(nil)
	ctx.Flags.ConfigPath = "config.yml"
	ctx.Env.Directory = "_testdata/projectdir"
	go func() {
		eventChan <- file.Event{Op: file.Update, Path: "assets/app.js"}
		signalChan <- os.Interrupt
		eventChan <- file.Event{Op: file.Remove, Path: "assets/app.js"}
	}()
	notifier = new(testAdapter)
	notifier.On("notify", ctx, "assets/app.js")
	err = watch(ctx, eventChan, signalChan, notifier)
	assert.Nil(t, err)
	assert.Contains(t, stdOut.String(), "Watching for file changes")
	assert.Contains(t, stdOut.String(), "processing assets/app.js")
	assert.Contains(t, stdOut.String(), "Updated assets/app.js")
	assert.NotContains(t, stdOut.String(), "Deleted assets/app.js")
	notifier.AssertExpectations(t)
}

func TestPerform(t *testing.T) {
	key := "assets/app.js"

	ctx, m, _, _, se := createTestCtx()
	perform(ctx, "bad", file.Update, "")
	assert.Contains(t, se.String(), "readAsset: ")
	m.AssertExpectations(t)

	ctx, m, _, _, se = createTestCtx()
	ctx.Env.Directory = "_testdata/projectdir"
	m.On("UpdateAsset", shopify.Asset{Key: key, Checksum: "d41d8cd98f00b204e9800998ecf8427e"}, "").Return(fmt.Errorf("shopify says no update"), "")
	perform(ctx, key, file.Update, "")
	assert.Contains(t, se.String(), "shopify says no update")
	m.AssertExpectations(t)

	ctx, m, _, so, _ := createTestCtx()
	ctx.Env.Directory = "_testdata/projectdir"
	m.On("UpdateAsset", shopify.Asset{Key: key, Checksum: "d41d8cd98f00b204e9800998ecf8427e"}, "").Return(nil)
	perform(ctx, key, file.Update, "")
	assert.NotContains(t, so.String(), "Updated")
	m.AssertExpectations(t)

	ctx, m, _, so, _ = createTestCtx()
	ctx.Env.Directory = "_testdata/projectdir"
	ctx.Flags.Verbose = true
	m.On("UpdateAsset", shopify.Asset{Key: key, Checksum: "d41d8cd98f00b204e9800998ecf8427e"}, "").Return(nil)
	perform(ctx, key, file.Update, "")
	assert.Contains(t, so.String(), "Updated")
	m.AssertExpectations(t)

	ctx, m, _, so, se = createTestCtx()
	m.On("DeleteAsset", mock.MatchedBy(func(a shopify.Asset) bool { return a.Key == "good" })).Return(nil)
	m.On("DeleteAsset", mock.MatchedBy(func(a shopify.Asset) bool { return a.Key == "bad" })).Return(fmt.Errorf("shopify says no update"))

	perform(ctx, "bad", file.Remove, "")
	assert.Contains(t, se.String(), "shopify says no update")

	perform(ctx, "good", file.Remove, "")
	assert.NotContains(t, so.String(), "Deleted")

	ctx.Flags.Verbose = true
	perform(ctx, "good", file.Remove, "")
	assert.Contains(t, so.String(), "Deleted")

	m.AssertExpectations(t)
}
