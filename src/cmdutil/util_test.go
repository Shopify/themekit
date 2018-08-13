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

	e = &env.Env{Proxy: "http://localhost:3000"}
	client = new(mocks.ShopifyClient)
	client.On("GetShop").Return(shopify.Shop{}, nil)
	client.On("Themes").Return([]shopify.Theme{{ID: 65443, Role: "unpublished"}, {ID: 1234, Role: "main"}}, nil)
	_, err = createCtx(factory, env.Conf{}, e, Flags{DisableIgnore: true}, []string{}, nil)
	assert.Nil(t, err)
	assert.Equal(t, e.ThemeID, "1234")
}

func TestCtx_StartProgress(t *testing.T) {
	ctx := Ctx{Env: &env.Env{}, Flags: Flags{}, progress: mpb.New(nil)}
	ctx.StartProgress(6)
	assert.NotNil(t, ctx.Bar)

	ctx = Ctx{Env: &env.Env{}, Flags: Flags{Verbose: true}, progress: mpb.New(nil)}
	ctx.StartProgress(6)
	assert.Nil(t, ctx.Bar)
}

func TestCtx_DoneTask(t *testing.T) {
	ctx := Ctx{Env: &env.Env{}, Flags: Flags{}, progress: mpb.New(nil)}
	assert.NotPanics(t, ctx.DoneTask)
	ctx.StartProgress(6)
	assert.NotNil(t, ctx.Bar)
	assert.Equal(t, ctx.Bar.Current(), int64(0))
	ctx.DoneTask()
	assert.Equal(t, ctx.Bar.Current(), int64(1))
}

func TestGenerateContexts(t *testing.T) {
	factory := func(*env.Env) (shopifyClient, error) { return nil, nil }
	_, err := generateContexts(factory, nil, Flags{}, []string{})
	assert.EqualError(t, err, "Could not find config file at ")

	client := new(mocks.ShopifyClient)
	factory = func(*env.Env) (shopifyClient, error) { return client, nil }
	client.On("GetShop").Return(shopify.Shop{}, nil)
	client.On("Themes").Return([]shopify.Theme{}, nil)
	ctxs, err := generateContexts(factory, nil, Flags{ConfigPath: "_testdata/config.yml"}, []string{})
	assert.Nil(t, err)
	assert.Equal(t, len(ctxs), 1)

	client = new(mocks.ShopifyClient)
	factory = func(*env.Env) (shopifyClient, error) { return client, nil }
	_, err = generateContexts(factory, nil, Flags{ConfigPath: "_testdata/config.yml", Environments: stringArgArray{[]string{"nope"}}}, []string{})
	assert.EqualError(t, err, "Could not load any valid environments")

	client = new(mocks.ShopifyClient)
	factory = func(*env.Env) (shopifyClient, error) { return client, nil }
	_, err = generateContexts(factory, nil, Flags{ConfigPath: "_testdata/invalid_config.yml", Environments: stringArgArray{[]string{"other"}}}, []string{})
	assert.EqualError(t, err, "invalid config invalid environment []: (invalid store domain must end in '.myshopify.com',missing password)")

	client = new(mocks.ShopifyClient)
	factory = func(*env.Env) (shopifyClient, error) { return client, fmt.Errorf("not today") }
	_, err = generateContexts(factory, nil, Flags{ConfigPath: "_testdata/config.yml"}, []string{})
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
		NotifyFile:   "n",
		IgnoredFiles: stringArgArray{[]string{"i"}},
		Ignores:      stringArgArray{[]string{"c"}},
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

func TestShouldUseEnvironment(t *testing.T) {
	testcases := []struct {
		envs, should, shouldNot []string
		all                     bool
	}{
		{envs: []string{}, should: []string{"development"}, shouldNot: []string{"production"}},
		{envs: []string{"production"}, should: []string{"production"}, shouldNot: []string{"development"}},
		{envs: []string{"p*", "other"}, should: []string{"production", "prod", "puddle", "other"}, shouldNot: []string{"development"}},
		{all: true, envs: []string{}, should: []string{"nope"}, shouldNot: []string{}},
	}

	for i, testcase := range testcases {
		flags := Flags{
			Environments: stringArgArray{testcase.envs},
			AllEnvs:      testcase.all,
		}

		for _, name := range testcase.should {
			assert.True(t, shouldUseEnvironment(flags, name), fmt.Sprintf("testcase number %v name: %v", i, name))
		}

		for _, name := range testcase.shouldNot {
			assert.False(t, shouldUseEnvironment(flags, name), fmt.Sprintf("testcase number %v name: %v", i, name))
		}
	}
}

func TestForEachClient(t *testing.T) {
	gandalfErr := fmt.Errorf("you shall not pass")
	safeHandler := func(Ctx) error { return nil }
	errHandler := func(Ctx) error { return gandalfErr }

	factory := func(*env.Env) (shopifyClient, error) { return nil, nil }
	err := forEachClient(factory, Flags{}, []string{}, safeHandler)
	assert.EqualError(t, err, "Could not find config file at ")

	client := new(mocks.ShopifyClient)
	factory = func(*env.Env) (shopifyClient, error) { return client, nil }
	client.On("GetShop").Return(shopify.Shop{}, nil)
	client.On("Themes").Return([]shopify.Theme{}, nil)
	err = forEachClient(factory, Flags{ConfigPath: "_testdata/config.yml"}, []string{}, safeHandler)
	assert.Nil(t, err)

	client = new(mocks.ShopifyClient)
	factory = func(*env.Env) (shopifyClient, error) { return client, nil }
	client.On("GetShop").Return(shopify.Shop{}, nil)
	client.On("Themes").Return([]shopify.Theme{}, nil)
	err = forEachClient(factory, Flags{ConfigPath: "_testdata/config.yml"}, []string{}, errHandler)
	assert.EqualError(t, err, gandalfErr.Error())

	count := 0
	handler := func(Ctx) error {
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
	err = forEachClient(factory, Flags{ConfigPath: "_testdata/config.yml"}, []string{}, handler)
	assert.EqualError(t, err, "nope not at all")
	assert.Equal(t, 2, count)
}

func TestForSingleClient(t *testing.T) {
	gandalfErr := fmt.Errorf("you shall not pass")
	safeHandler := func(Ctx) error { return nil }
	errHandler := func(Ctx) error { return gandalfErr }

	factory := func(*env.Env) (shopifyClient, error) { return nil, nil }
	err := forSingleClient(factory, Flags{}, []string{}, safeHandler)
	assert.EqualError(t, err, "Could not find config file at ")

	client := new(mocks.ShopifyClient)
	factory = func(*env.Env) (shopifyClient, error) { return client, nil }
	client.On("GetShop").Return(shopify.Shop{}, nil)
	client.On("Themes").Return([]shopify.Theme{}, nil)
	err = forSingleClient(factory, Flags{ConfigPath: "_testdata/config.yml"}, []string{}, safeHandler)
	assert.Nil(t, err)

	client = new(mocks.ShopifyClient)
	factory = func(*env.Env) (shopifyClient, error) { return client, nil }
	client.On("GetShop").Return(shopify.Shop{}, nil)
	client.On("Themes").Return([]shopify.Theme{}, nil)
	err = forSingleClient(factory, Flags{ConfigPath: "_testdata/config.yml", Environments: stringArgArray{[]string{"*"}}}, []string{}, safeHandler)
	assert.EqualError(t, err, "more than one environment specified for a single environment command")

	client = new(mocks.ShopifyClient)
	factory = func(*env.Env) (shopifyClient, error) { return client, nil }
	client.On("GetShop").Return(shopify.Shop{}, nil)
	client.On("Themes").Return([]shopify.Theme{}, nil)
	err = forSingleClient(factory, Flags{ConfigPath: "_testdata/config.yml"}, []string{}, errHandler)
	assert.EqualError(t, err, gandalfErr.Error())

	count := 0
	handler := func(Ctx) error {
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
	err = forSingleClient(factory, Flags{ConfigPath: "_testdata/config.yml"}, []string{}, handler)
	assert.EqualError(t, err, "nope not at all")
	assert.Equal(t, 2, count)
}

func TestForDefaultClient(t *testing.T) {
	gandalfErr := fmt.Errorf("you shall not pass")
	safeHandler := func(Ctx) error { return nil }
	errHandler := func(Ctx) error { return gandalfErr }

	factory := func(*env.Env) (shopifyClient, error) { return nil, nil }
	err := forDefaultClient(factory, Flags{}, []string{}, safeHandler)
	assert.EqualError(t, err, "invalid environment [development]: (missing store domain,missing password)")

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
	err = forDefaultClient(factory, Flags{Domain: "shop.myshopify.com", Password: "123"}, []string{}, safeHandler)
	assert.Nil(t, err)

	client = new(mocks.ShopifyClient)
	factory = func(*env.Env) (shopifyClient, error) { return client, fmt.Errorf("server err") }
	err = forDefaultClient(factory, Flags{Domain: "shop.myshopify.com", Password: "123"}, []string{}, safeHandler)
	assert.EqualError(t, err, "server err")

	client = new(mocks.ShopifyClient)
	factory = func(*env.Env) (shopifyClient, error) { return client, nil }
	client.On("GetShop").Return(shopify.Shop{}, nil)
	client.On("Themes").Return([]shopify.Theme{}, nil)
	err = forDefaultClient(factory, Flags{Domain: "shop.myshopify.com", Password: "123"}, []string{}, errHandler)
	assert.EqualError(t, err, gandalfErr.Error())
}

func createTestCtx() (ctx Ctx, m *mocks.ShopifyClient, stdOut, stdErr *bytes.Buffer) {
	m = new(mocks.ShopifyClient)
	stdOut, stdErr = bytes.NewBufferString(""), bytes.NewBufferString("")
	ctx = Ctx{
		Conf:     &env.Conf{},
		Client:   m,
		Env:      &env.Env{},
		Flags:    Flags{},
		Log:      log.New(stdOut, "", 0),
		ErrLog:   log.New(stdErr, "", 0),
		progress: mpb.New(nil),
	}
	return
}
