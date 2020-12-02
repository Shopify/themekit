package cmdutil

import (
	"bytes"
	"fmt"
	"log"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vbauerster/mpb"

	"github.com/Shopify/themekit/src/cmdutil/_mocks"
	"github.com/Shopify/themekit/src/env"
	"github.com/Shopify/themekit/src/file"
	"github.com/Shopify/themekit/src/shopify"
)

func TestCreateCtx(t *testing.T) {
	e := &env.Env{Domain: "this is not a url%@#$@#"}
	client := new(mocks.ShopifyClient)
	factory := func(*env.Env) (shopifyClient, error) { return client, nil }
	client.On("GetShop").Return(shopify.Shop{}, shopify.ErrShopDomainNotFound)
	_, err := createCtx(factory, env.Conf{}, e, Flags{}, []string{}, nil)
	if assert.NotNil(t, err) {
		assert.Contains(t, err.Error(), "invalid domain")
	}

	e = &env.Env{Domain: "this is not a url%@#$@#"}
	client = new(mocks.ShopifyClient)
	factory = func(*env.Env) (shopifyClient, error) { return client, nil }
	client.On("GetShop").Return(shopify.Shop{}, fmt.Errorf("This is bad"))
	_, err = createCtx(factory, env.Conf{}, e, Flags{}, []string{}, nil)
	if assert.NotNil(t, err) {
		assert.Contains(t, err.Error(), "This is bad")
	}

	client = new(mocks.ShopifyClient)
	client.On("GetShop").Return(shopify.Shop{}, nil)
	client.On("Themes").Return([]shopify.Theme{}, nil)
	badFactory := func(*env.Env) (shopifyClient, error) { return nil, fmt.Errorf("no such file or directory") }
	_, err = createCtx(badFactory, env.Conf{}, &env.Env{}, Flags{}, []string{}, nil)
	if assert.NotNil(t, err) {
		assert.Contains(t, err.Error(), "no such file or directory")
	}

	client = new(mocks.ShopifyClient)
	client.On("GetShop").Return(shopify.Shop{}, nil)
	client.On("Themes").Return([]shopify.Theme{}, fmt.Errorf("[API] Invalid API key or access token (unrecognized login or wrong password)"))
	_, err = createCtx(factory, env.Conf{}, &env.Env{}, Flags{}, []string{}, nil)
	if assert.NotNil(t, err) {
		assert.Contains(t, err.Error(), "[API] Invalid API key or access token (unrecognized login or wrong password)")
	}

	e = &env.Env{ThemeID: "1234", Proxy: "http://localhost:3000"}
	client = new(mocks.ShopifyClient)
	client.On("GetShop").Return(shopify.Shop{}, nil)
	client.On("Themes").Return([]shopify.Theme{{ID: 65443, Role: "unpublished"}, {ID: 1234, Role: "main"}}, nil)
	_, err = createCtx(factory, env.Conf{}, e, Flags{DisableIgnore: true}, []string{}, nil)
	assert.Equal(t, ErrLiveTheme, err)
	assert.Equal(t, e.ThemeID, "1234")
}

func TestCtx_StartProgress(t *testing.T) {
	ctx := Ctx{Env: &env.Env{}, Flags: Flags{Verbose: true}, progress: mpb.New(nil)}
	ctx.StartProgress(6)
	assert.Nil(t, ctx.Bar)
	ctx.Flags.Verbose = false
	ctx.StartProgress(6)
	assert.NotNil(t, ctx.Bar)
}

func TestCtx_Err(t *testing.T) {
	stdErr := bytes.NewBufferString("")
	ctx := Ctx{Env: &env.Env{}, Flags: Flags{}, progress: mpb.New(nil), ErrLog: log.New(stdErr, "", 0)}

	ctx.Err("[%s] this is err", "Development")
	assert.Contains(t, stdErr.String(), "[Development] this is err")

	ctx.StartProgress(6)
	assert.NotNil(t, ctx.Bar)

	ctx.Err("[%s] this is err", "production")
	assert.Equal(t, ctx.summary.errors[1], "[production] this is err")
	assert.NotContains(t, stdErr.String(), "[production] this is err")
}

func TestCtx_DoneTask(t *testing.T) {
	ctx := Ctx{Env: &env.Env{}, Flags: Flags{}, progress: mpb.New(nil)}
	assert.NotPanics(t, func() {
		ctx.DoneTask(file.Update)
	})
	ctx.StartProgress(6)
	assert.NotNil(t, ctx.Bar)
	assert.Equal(t, ctx.Bar.Current(), int64(0))
	ctx.DoneTask(file.Update)
	assert.Equal(t, ctx.Bar.Current(), int64(1))
}

func TestGenerateContexts(t *testing.T) {
	factory := func(*env.Env) (shopifyClient, error) { return nil, nil }
	_, err := generateContexts(factory, nil, Flags{Environments: []string{"development"}}, []string{})
	assert.EqualError(t, err, "invalid environment [development]: (missing theme_id,missing store domain,missing password)")

	client := new(mocks.ShopifyClient)
	factory = func(*env.Env) (shopifyClient, error) { return client, nil }
	client.On("GetShop").Return(shopify.Shop{}, nil)
	client.On("Themes").Return([]shopify.Theme{}, nil)
	ctxs, err := generateContexts(factory, nil, Flags{Environments: []string{"development"}, ConfigPath: "_testdata/config.yml"}, []string{})
	assert.Nil(t, err)
	assert.Equal(t, len(ctxs), 1)

	client = new(mocks.ShopifyClient)
	factory = func(*env.Env) (shopifyClient, error) { return client, nil }
	_, err = generateContexts(factory, nil, Flags{ConfigPath: "_testdata/config.yml", Environments: []string{"nope"}}, []string{})
	assert.EqualError(t, err, "invalid environment [nope]: (missing theme_id,missing store domain,missing password)")

	client = new(mocks.ShopifyClient)
	factory = func(*env.Env) (shopifyClient, error) { return client, fmt.Errorf("not today") }
	_, err = generateContexts(factory, nil, Flags{Environments: []string{"development"}, ConfigPath: "_testdata/config.yml"}, []string{})
	assert.EqualError(t, err, "not today")
}

func TestGetFlagEnv(t *testing.T) {
	flags := Flags{
		Directory:    "d",
		Password:     "p",
		ThemeID:      "t",
		Domain:       "o",
		Proxy:        "r",
		Timeout:      1,
		Notify:       "n",
		IgnoredFiles: []string{"i"},
		Ignores:      []string{"c"},
	}

	e := env.Env{
		Directory:    "d",
		Password:     "p",
		ThemeID:      "t",
		Domain:       "o",
		Proxy:        "r",
		Timeout:      1,
		Notify:       "n",
		IgnoredFiles: []string{"i"},
		Ignores:      []string{"c"},
	}

	assert.Equal(t, e, getFlagEnv(flags))

	flags.DisableIgnore = true
	assert.NotEqual(t, e, getFlagEnv(flags))

	e = env.Env{
		Directory: "d",
		Password:  "p",
		ThemeID:   "t",
		Domain:    "o",
		Proxy:     "r",
		Timeout:   1,
		Notify:    "n",
	}

	assert.Equal(t, e, getFlagEnv(flags))
}

func TestExpandEnvironments(t *testing.T) {
	testcases := []struct {
		envs, conf, expected []string
		all                  bool
	}{
		{envs: []string{}, conf: []string{"development"}, expected: []string{}},
		{envs: []string{"production"}, conf: []string{"production", "foob"}, expected: []string{"production"}},
		{envs: []string{"production"}, conf: []string{}, expected: []string{"production"}},
		{envs: []string{"p*"}, conf: []string{"production", "prod", "puddle", "other"}, expected: []string{"production", "prod", "puddle"}},
		{all: true, envs: []string{}, conf: []string{"dev", "prod", "test"}, expected: []string{"dev", "prod", "test"}},
	}

	for _, testcase := range testcases {
		flags := Flags{
			Environments: testcase.envs,
			AllEnvs:      testcase.all,
		}
		envs := map[string]*env.Env{}
		for _, name := range testcase.conf {
			envs[name] = nil
		}
		got := expandEnvironments(flags, envs)
		assert.ElementsMatch(t, testcase.expected, got)
	}
}

func TestForEachClient(t *testing.T) {
	gandalfErr := fmt.Errorf("you shall not pass")
	safeHandler := func(*Ctx) error { return nil }
	errHandler := func(*Ctx) error { return gandalfErr }

	client := new(mocks.ShopifyClient)
	factory := func(*env.Env) (shopifyClient, error) { return client, nil }
	client.On("GetShop").Return(shopify.Shop{}, nil)
	client.On("Themes").Return([]shopify.Theme{}, nil)
	err := forEachClient(factory, Flags{ConfigPath: "_testdata/config.yml"}, []string{}, safeHandler)
	assert.Nil(t, err)

	client = new(mocks.ShopifyClient)
	factory = func(*env.Env) (shopifyClient, error) { return client, nil }
	client.On("GetShop").Return(shopify.Shop{}, nil)
	client.On("Themes").Return([]shopify.Theme{}, nil)
	err = forEachClient(factory, Flags{Environments: []string{"development"}, ConfigPath: "_testdata/config.yml"}, []string{}, errHandler)
	assert.EqualError(t, err, gandalfErr.Error())

	count := 0
	handler := func(*Ctx) error {
		count++
		if count == 1 {
			return ErrReload
		}
		return fmt.Errorf("nope not at all")
	}
	client = new(mocks.ShopifyClient)
	factory = func(*env.Env) (shopifyClient, error) { return client, nil }
	client.On("GetShop").Return(shopify.Shop{}, nil)
	client.On("Themes").Return([]shopify.Theme{}, nil)
	err = forEachClient(factory, Flags{Environments: []string{"development"}, ConfigPath: "_testdata/config.yml"}, []string{}, handler)
	assert.EqualError(t, err, "nope not at all")
	assert.Equal(t, 2, count)

	stdErr := bytes.NewBufferString("")
	handler = func(ctx *Ctx) error {
		ctx.ErrLog = log.New(stdErr, "", 0)
		ctx.StartProgress(1)
		ctx.Err("oopsy")
		ctx.DoneTask(file.Skip)
		return nil
	}
	client = new(mocks.ShopifyClient)
	factory = func(*env.Env) (shopifyClient, error) { return client, nil }
	client.On("GetShop").Return(shopify.Shop{}, nil)
	client.On("Themes").Return([]shopify.Theme{}, nil)
	err = forEachClient(factory, Flags{Environments: []string{"development"}, ConfigPath: "_testdata/config.yml"}, []string{}, handler)
	assert.Equal(t, ErrDuringRuntime, err)
	assert.Contains(t, stdErr.String(), "Errors encountered: ")
}

func TestForSingleClient(t *testing.T) {
	gandalfErr := fmt.Errorf("you shall not pass")
	safeHandler := func(*Ctx) error { return nil }
	errHandler := func(*Ctx) error { return gandalfErr }

	client := new(mocks.ShopifyClient)
	factory := func(*env.Env) (shopifyClient, error) { return client, nil }
	client.On("GetShop").Return(shopify.Shop{}, nil)
	client.On("Themes").Return([]shopify.Theme{}, nil)
	err := forSingleClient(factory, Flags{Environments: []string{"development"}, ConfigPath: "_testdata/config.yml"}, []string{}, safeHandler)
	assert.Nil(t, err)

	client = new(mocks.ShopifyClient)
	factory = func(*env.Env) (shopifyClient, error) { return client, nil }
	client.On("GetShop").Return(shopify.Shop{}, nil)
	client.On("Themes").Return([]shopify.Theme{}, nil)
	err = forSingleClient(factory, Flags{ConfigPath: "_testdata/config.yml", Environments: []string{"*"}}, []string{}, safeHandler)
	assert.EqualError(t, err, "more than one environment specified for a single environment command")

	client = new(mocks.ShopifyClient)
	factory = func(*env.Env) (shopifyClient, error) { return client, nil }
	client.On("GetShop").Return(shopify.Shop{}, nil)
	client.On("Themes").Return([]shopify.Theme{}, nil)
	err = forSingleClient(factory, Flags{Environments: []string{"development"}, ConfigPath: "_testdata/config.yml"}, []string{}, errHandler)
	assert.EqualError(t, err, gandalfErr.Error())

	count := 0
	handler := func(*Ctx) error {
		count++
		if count == 1 {
			return ErrReload
		}
		return fmt.Errorf("nope not at all")
	}
	client = new(mocks.ShopifyClient)
	factory = func(*env.Env) (shopifyClient, error) { return client, nil }
	client.On("GetShop").Return(shopify.Shop{}, nil)
	client.On("Themes").Return([]shopify.Theme{}, nil)
	err = forSingleClient(factory, Flags{Environments: []string{"development"}, ConfigPath: "_testdata/config.yml"}, []string{}, handler)
	assert.EqualError(t, err, "nope not at all")
	assert.Equal(t, 2, count)

	stdErr := bytes.NewBufferString("")
	handler = func(ctx *Ctx) error {
		ctx.ErrLog = log.New(stdErr, "", 0)
		ctx.StartProgress(1)
		ctx.Err("oopsy")
		ctx.DoneTask(file.Skip)
		return nil
	}
	client = new(mocks.ShopifyClient)
	factory = func(*env.Env) (shopifyClient, error) { return client, nil }
	client.On("GetShop").Return(shopify.Shop{}, nil)
	client.On("Themes").Return([]shopify.Theme{}, nil)
	err = forSingleClient(factory, Flags{Environments: []string{"development"}, ConfigPath: "_testdata/config.yml"}, []string{}, handler)
	assert.Equal(t, ErrDuringRuntime, err)
	assert.Contains(t, stdErr.String(), "Errors encountered")
}

func TestForDefaultClient(t *testing.T) {
	gandalfErr := fmt.Errorf("you shall not pass")
	safeHandler := func(*Ctx) error { return nil }
	errHandler := func(*Ctx) error { return gandalfErr }

	factory := func(*env.Env) (shopifyClient, error) { return nil, nil }
	err := forDefaultClient(factory, Flags{}, []string{}, safeHandler)
	assert.EqualError(t, err, "invalid environment [development]: (missing theme_id,missing store domain,missing password)")

	client := new(mocks.ShopifyClient)
	factory = func(*env.Env) (shopifyClient, error) { return client, nil }
	client.On("GetShop").Return(shopify.Shop{}, nil)
	client.On("Themes").Return([]shopify.Theme{}, nil)
	err = forDefaultClient(factory, Flags{ConfigPath: "_testdata/config.yml"}, []string{}, safeHandler)
	assert.Nil(t, err)

	client = new(mocks.ShopifyClient)
	factory = func(*env.Env) (shopifyClient, error) { return client, nil }
	client.On("GetShop").Return(shopify.Shop{}, nil)
	client.On("Themes").Return([]shopify.Theme{}, nil)
	err = forDefaultClient(factory, Flags{Domain: "shop.myshopify.com", Password: "123", ThemeID: "123"}, []string{}, safeHandler)
	assert.Nil(t, err)

	client = new(mocks.ShopifyClient)
	factory = func(*env.Env) (shopifyClient, error) { return client, fmt.Errorf("server err") }
	err = forDefaultClient(factory, Flags{Domain: "shop.myshopify.com", Password: "123", ThemeID: "123"}, []string{}, safeHandler)
	assert.EqualError(t, err, "server err")

	client = new(mocks.ShopifyClient)
	factory = func(*env.Env) (shopifyClient, error) { return client, nil }
	client.On("GetShop").Return(shopify.Shop{}, nil)
	client.On("Themes").Return([]shopify.Theme{}, nil)
	err = forDefaultClient(factory, Flags{Domain: "shop.myshopify.com", Password: "123", ThemeID: "123"}, []string{}, errHandler)
	assert.EqualError(t, err, gandalfErr.Error())

	stdErr := bytes.NewBufferString("")
	handler := func(ctx *Ctx) error {
		ctx.ErrLog = log.New(stdErr, "", 0)
		ctx.StartProgress(1)
		ctx.Err("oopsy")
		ctx.DoneTask(file.Skip)
		return nil
	}
	client = new(mocks.ShopifyClient)
	factory = func(*env.Env) (shopifyClient, error) { return client, nil }
	client.On("GetShop").Return(shopify.Shop{}, nil)
	client.On("Themes").Return([]shopify.Theme{}, nil)
	forDefaultClient(factory, Flags{ConfigPath: "_testdata/config.yml"}, []string{}, handler)
	assert.Equal(t, gandalfErr, err)
	assert.Contains(t, stdErr.String(), "Errors encountered: ")
}
