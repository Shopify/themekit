package cmd

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Shopify/themekit/src/cmdutil"
	"github.com/Shopify/themekit/src/env"
	"github.com/Shopify/themekit/src/shopify"
)

func TestNewTheme(t *testing.T) {
	name, url := "name", "https://download.com/1.2.4.zip"

	ctx, client, conf, _, _ := createTestCtx()
	client.On("CreateNewTheme", name, url).Return(shopify.Theme{}, nil)
	client.On("GetInfo").Return(shopify.Theme{Previewable: true}, nil)
	client.On("GetAllAssets").Return([]string{}, nil)
	conf.On("Set", "development", env.Env{}).Return(nil, nil)
	conf.On("Save").Return(nil)
	err := newTheme(ctx, name, url)
	assert.Error(t, err)

	ctx, client, _, _, _ = createTestCtx()
	client.On("CreateNewTheme", name, url).Return(shopify.Theme{}, fmt.Errorf("can't create theme"))
	err = newTheme(ctx, name, url)
	if assert.NotNil(t, err) {
		assert.Contains(t, err.Error(), "can't create theme")
	}

	ctx, client, conf, _, _ = createTestCtx()
	client.On("CreateNewTheme", name, url).Return(shopify.Theme{}, nil)
	conf.On("Set", "development", env.Env{}).Return(nil, fmt.Errorf("cant set config"))
	err = newTheme(ctx, name, url)
	if assert.NotNil(t, err) {
		assert.Contains(t, err.Error(), "cant set config")
	}

	ctx, client, conf, _, _ = createTestCtx()
	client.On("CreateNewTheme", name, url).Return(shopify.Theme{}, nil)
	client.On("GetInfo").Return(shopify.Theme{}, fmt.Errorf("oh no"))
	client.On("GetAllAssets").Return([]string{}, nil)
	conf.On("Set", "development", env.Env{}).Return(nil, nil)
	conf.On("Save").Return(nil)
	err = newTheme(ctx, name, url)
	if assert.NotNil(t, err) {
		assert.Contains(t, err.Error(), "oh no")
	}
}

func TestGetNewThemeDetails(t *testing.T) {
	getVerGood := func(string) (string, error) { return "https://download.com", nil }
	getVerBad := func(string) (string, error) { return "", fmt.Errorf("cant fetch releases") }

	testcases := []struct {
		in, iu, on, ou, err string
		handler             func(string) (string, error)
	}{
		{on: "preTimber-latest", ou: "https://download.com", handler: getVerGood},
		{on: "preTimber-latest", ou: "", handler: getVerBad, err: "cant fetch releases"},
		{iu: "https://download.com/1.2.4.zip", on: "pre1.2.4", ou: "https://download.com/1.2.4.zip", handler: getVerGood},
		{in: "name", iu: "https://download.com/1.2.4.zip", on: "name", ou: "https://download.com/1.2.4.zip", handler: getVerBad},
	}

	for _, testcase := range testcases {
		flags := cmdutil.Flags{
			Name:    testcase.in,
			URL:     testcase.iu,
			Version: "latest",
			Prefix:  "pre",
		}
		name, url, err := getNewThemeDetails(flags, testcase.handler)
		assert.Equal(t, testcase.on, name)
		assert.Equal(t, testcase.ou, url)
		if testcase.err == "" {
			assert.Nil(t, err)
		} else if assert.NotNil(t, err) {
			assert.Contains(t, err.Error(), testcase.err)
		}
	}
}
