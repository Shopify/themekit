package cmd

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Shopify/themekit/src/cmdutil"
	"github.com/Shopify/themekit/src/env"
	"github.com/Shopify/themekit/src/shopify"
)

func TestNewTheme(t *testing.T) {
	name := "name"

	ctx, client, conf, _, _ := createTestCtx()
	ctx.Flags.Name = name
	client.On("CreateNewTheme", name).Return(shopify.Theme{ID: 42}, nil)
	conf.On("Set", "development", env.Env{ThemeID: "42"}).Return(nil, nil)
	conf.On("Save").Return(nil)
	client.On("GetAllAssets").Return([]shopify.Asset{}, nil)
	err := newTheme(ctx, func(ctx *cmdutil.Ctx) error { return nil })
	assert.Error(t, err)

	ctx, client, _, _, _ = createTestCtx()
	ctx.Flags.Name = name
	client.On("CreateNewTheme", name).Return(shopify.Theme{}, fmt.Errorf("can't create theme"))
	err = newTheme(ctx, func(ctx *cmdutil.Ctx) error { return nil })
	if assert.NotNil(t, err) {
		assert.Contains(t, err.Error(), "can't create theme")
	}

	ctx, client, conf, _, _ = createTestCtx()
	ctx.Flags.Name = name
	client.On("CreateNewTheme", name).Return(shopify.Theme{ID: 44}, nil)
	conf.On("Set", "development", env.Env{ThemeID: "44"}).Return(nil, fmt.Errorf("cant set config"))
	err = newTheme(ctx, func(ctx *cmdutil.Ctx) error { return nil })
	if assert.NotNil(t, err) {
		assert.Contains(t, err.Error(), "cant set config")
	}

	ctx, client, conf, _, _ = createTestCtx()
	ctx.Flags.Name = name
	client.On("CreateNewTheme", name).Return(shopify.Theme{ID: 48}, nil)
	conf.On("Set", "development", env.Env{ThemeID: "48"}).Return(nil, nil)
	conf.On("Save").Return(nil)
	err = newTheme(ctx, func(ctx *cmdutil.Ctx) error { return errors.New("oh no") })
	if assert.NotNil(t, err) {
		assert.Contains(t, err.Error(), "oh no")
	}
}
