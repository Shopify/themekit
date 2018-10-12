package cmd

import (
	"bytes"
	"log"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Shopify/themekit/src/cmdutil"
	"github.com/Shopify/themekit/src/cmdutil/_mocks"
	"github.com/Shopify/themekit/src/env"
	"github.com/Shopify/themekit/src/shopify"
)

func TestRemove(t *testing.T) {
	testcases := []struct {
		args, err string
		readonly  bool
	}{
		{args: filepath.Join("templates", "layout.liquid")},
		{args: filepath.Join("templates", "layout.liquid"), readonly: true, err: "environment is readonly"},
		{err: "please specify file(s) to be removed"},
	}

	for _, testcase := range testcases {
		ctx, client, _, _, _ := createTestCtx()
		if testcase.args != "" {
			ctx.Args = []string{testcase.args}
		}
		ctx.Env.ReadOnly = testcase.readonly

		client.On("DeleteAsset", shopify.Asset{Key: testcase.args}).Return(nil)

		err := remove(ctx, func(path string) error {
			assert.Equal(t, path, testcase.args)
			return nil
		})

		if testcase.err == "" {
			assert.Nil(t, err)
		} else if assert.NotNil(t, err, testcase.err) {
			assert.Contains(t, err.Error(), testcase.err)
		}
	}
}

func createTestCtx() (ctx *cmdutil.Ctx, client *mocks.ShopifyClient, conf *mocks.Config, stdOut, stdErr *bytes.Buffer) {
	client = new(mocks.ShopifyClient)
	conf = new(mocks.Config)
	stdOut, stdErr = bytes.NewBufferString(""), bytes.NewBufferString("")
	ctx = &cmdutil.Ctx{
		Conf:   conf,
		Client: client,
		Env:    &env.Env{},
		Flags: cmdutil.Flags{
			Environments: []string{"development"},
		},
		Log:    log.New(stdOut, "", 0),
		ErrLog: log.New(stdErr, "", 0),
	}
	return
}
