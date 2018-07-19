package cmdutil

import (
	"bytes"
	"fmt"
	"log"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/vbauerster/mpb"

	"github.com/Shopify/themekit/src/cmdutil/_mocks"
	"github.com/Shopify/themekit/src/env"
	"github.com/Shopify/themekit/src/shopify"
)

func TestCreateCtx(t *testing.T) {
	e := &env.Env{Domain: "this is not a url%@#$@#"}
	_, err := createCtx(env.Conf{}, e, Flags{}, []string{}, nil)
	if assert.NotNil(t, err) {
		assert.Contains(t, err.Error(), "invalid domain")
	}

	e = &env.Env{Ignores: []string{"notthere"}}
	_, err = createCtx(env.Conf{}, e, Flags{}, []string{}, nil)
	if assert.NotNil(t, err) {
		assert.Contains(t, err.Error(), "no such file or directory")
	}

	_, err = createCtx(env.Conf{}, &env.Env{}, Flags{}, []string{}, nil)
	assert.Nil(t, err)
}

func TestCtx_StartProgress(t *testing.T) {
	ctx, err := createCtx(env.Conf{}, &env.Env{}, Flags{}, []string{}, mpb.New(nil))
	assert.Nil(t, err)
	assert.Nil(t, ctx.Bar)
	ctx.StartProgress(6)
	assert.NotNil(t, ctx.Bar)

	ctx, err = createCtx(env.Conf{}, &env.Env{}, Flags{Verbose: true}, []string{}, mpb.New(nil))
	assert.Nil(t, err)
	assert.Nil(t, ctx.Bar)
	ctx.StartProgress(6)
	assert.Nil(t, ctx.Bar)
}

func TestCtx_DoneTask(t *testing.T) {
	ctx, err := createCtx(env.Conf{}, &env.Env{}, Flags{}, []string{}, mpb.New(nil))
	assert.Nil(t, err)
	assert.Nil(t, ctx.Bar)
	assert.NotPanics(t, ctx.DoneTask)
	ctx.StartProgress(6)
	assert.NotNil(t, ctx.Bar)
	assert.Equal(t, ctx.Bar.Current(), int64(0))
	ctx.DoneTask()
	assert.Equal(t, ctx.Bar.Current(), int64(1))
}

func TestGenerateContexts(t *testing.T) {
	testcases := []struct {
		flags Flags
		count int
		err   string
	}{
		{flags: Flags{}, err: "Could not find config file"},
		{flags: Flags{ConfigPath: "_testdata/config.yml"}, count: 1},
		{flags: Flags{ConfigPath: "_testdata/config.yml", Environments: stringArgArray{[]string{"nope"}}}, err: "Could not load any valid environments"},
		{flags: Flags{ConfigPath: "_testdata/invalid_config.yml", Environments: stringArgArray{[]string{"other"}}}, err: "invalid config"},
	}

	for _, testcase := range testcases {
		ctxs, err := generateContexts(nil, testcase.flags, []string{})
		if testcase.err == "" {
			assert.Nil(t, err)
			assert.Equal(t, testcase.count, len(ctxs))
		} else if assert.NotNil(t, err) {
			assert.Contains(t, err.Error(), testcase.err)
		}
	}
}

func TestGetFlagEnv(t *testing.T) {

	// if !flags.DisableIgnore {
	//	flagEnv.IgnoredFiles = flags.IgnoredFiles.Value()
	//	flagEnv.Ignores = flags.Ignores.Value()
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

	testcases := []struct {
		flags  Flags
		handle func(Ctx) error
		err    string
	}{
		{flags: Flags{}, handle: safeHandler, err: "Could not find config file"},
		{flags: Flags{ConfigPath: "_testdata/config.yml"}, handle: safeHandler},
		{flags: Flags{ConfigPath: "_testdata/config.yml"}, handle: errHandler, err: gandalfErr.Error()},
	}

	for _, testcase := range testcases {
		err := ForEachClient(testcase.flags, []string{}, testcase.handle)
		if testcase.err == "" {
			assert.Nil(t, err)
		} else if assert.NotNil(t, err) {
			assert.Contains(t, err.Error(), testcase.err)
		}
	}

	count := 0
	handler := func(Ctx) error {
		count++
		if count == 1 {
			return ErrReload
		}
		return fmt.Errorf("nope not at all")
	}
	err := ForEachClient(Flags{ConfigPath: "_testdata/config.yml"}, []string{}, handler)
	if assert.NotNil(t, err) {
		assert.Contains(t, err.Error(), "nope not at all")
	}
	assert.Equal(t, 2, count)
}

func TestForSingleClient(t *testing.T) {
	gandalfErr := fmt.Errorf("you shall not pass")
	safeHandler := func(Ctx) error { return nil }
	errHandler := func(Ctx) error { return gandalfErr }

	testcases := []struct {
		flags  Flags
		handle func(Ctx) error
		err    string
	}{
		{flags: Flags{}, handle: safeHandler, err: "Could not find config file"},
		{flags: Flags{ConfigPath: "_testdata/config.yml"}, handle: safeHandler},
		{flags: Flags{ConfigPath: "_testdata/config.yml"}, handle: errHandler, err: gandalfErr.Error()},
		{flags: Flags{ConfigPath: "_testdata/config.yml", Environments: stringArgArray{[]string{"*"}}}, handle: errHandler, err: "more than one environment specified for a single environment command"},
	}

	for _, testcase := range testcases {
		err := ForSingleClient(testcase.flags, []string{}, testcase.handle)
		if testcase.err == "" {
			assert.Nil(t, err)
		} else if assert.NotNil(t, err) {
			assert.Contains(t, err.Error(), testcase.err)
		}
	}

	count := 0
	handler := func(Ctx) error {
		count++
		if count == 1 {
			return ErrReload
		}
		return fmt.Errorf("nope not at all")
	}
	err := ForSingleClient(Flags{ConfigPath: "_testdata/config.yml"}, []string{}, handler)
	if assert.NotNil(t, err) {
		assert.Contains(t, err.Error(), "nope not at all")
	}
	assert.Equal(t, 2, count)
}

func TestForDefaultClient(t *testing.T) {
	gandalfErr := fmt.Errorf("you shall not pass")
	safeHandler := func(Ctx) error { return nil }
	errHandler := func(Ctx) error { return gandalfErr }

	testcases := []struct {
		flags  Flags
		handle func(Ctx) error
		err    string
	}{
		{flags: Flags{}, handle: safeHandler, err: "missing store domain"},
		{flags: Flags{ConfigPath: "_testdata/config.yml"}, handle: safeHandler},
		{flags: Flags{Domain: "shop.myshopify.com", Password: "123"}, handle: safeHandler},
		{flags: Flags{Domain: "shop.myshopify.com", Password: "123", Ignores: stringArgArray{[]string{"nothere"}}}, handle: safeHandler, err: "no such file"},
		{flags: Flags{ConfigPath: "_testdata/config.yml"}, handle: errHandler, err: gandalfErr.Error()},
	}

	for _, testcase := range testcases {
		err := ForDefaultClient(testcase.flags, []string{}, testcase.handle)
		if testcase.err == "" {
			assert.Nil(t, err)
		} else if assert.NotNil(t, err) {
			assert.Contains(t, err.Error(), testcase.err)
		}
	}
}

func TestUploadAsset(t *testing.T) {
	ctx, m, so, se := createTestCtx()

	m.On("UpdateAsset", mock.MatchedBy(func(a shopify.Asset) bool { return a.Key == "good" })).Return(nil)
	m.On("UpdateAsset", mock.MatchedBy(func(a shopify.Asset) bool { return a.Key == "bad" })).Return(fmt.Errorf("shopify says no update"))
	ctx.StartProgress(1)

	UploadAsset(ctx, shopify.Asset{Key: "bad"})
	assert.Equal(t, ctx.Bar.Current(), int64(1))
	assert.Contains(t, se.String(), "shopify says no update")

	UploadAsset(ctx, shopify.Asset{Key: "good"})
	assert.NotContains(t, so.String(), "Updated")

	ctx.Flags.Verbose = true
	UploadAsset(ctx, shopify.Asset{Key: "good"})
	assert.Contains(t, so.String(), "Updated")

	m.AssertExpectations(t)
}

func TestDeleteAsset(t *testing.T) {
	ctx, m, so, se := createTestCtx()

	m.On("DeleteAsset", mock.MatchedBy(func(a shopify.Asset) bool { return a.Key == "good" })).Return(nil)
	m.On("DeleteAsset", mock.MatchedBy(func(a shopify.Asset) bool { return a.Key == "bad" })).Return(fmt.Errorf("shopify says no update"))
	ctx.StartProgress(1)

	DeleteAsset(ctx, shopify.Asset{Key: "bad"})
	assert.Equal(t, ctx.Bar.Current(), int64(1))
	assert.Contains(t, se.String(), "shopify says no update")

	DeleteAsset(ctx, shopify.Asset{Key: "good"})
	assert.NotContains(t, so.String(), "Deleted")

	ctx.Flags.Verbose = true
	DeleteAsset(ctx, shopify.Asset{Key: "good"})
	assert.Contains(t, so.String(), "Deleted")

	m.AssertExpectations(t)
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
