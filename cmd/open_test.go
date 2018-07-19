package cmd

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOpen(t *testing.T) {
	ctx, _, _, _, _ := createTestCtx()
	ctx.Env.Domain = "my.test.domain"
	ctx.Env.ThemeID = "123"
	assert.Nil(t, preview(ctx, func(path string) error {
		assert.Equal(t, path, "https://my.test.domain?preview_theme_id=123")
		return nil
	}))

	ctx, _, _, _, _ = createTestCtx()
	ctx.Env.Domain = "my.test.domain"
	ctx.Env.ThemeID = "123"
	ctx.Flags.Edit = true
	assert.Nil(t, preview(ctx, func(path string) error {
		assert.Equal(t, path, "https://my.test.domain/admin/themes/123/editor")
		return nil
	}))

	ctx, _, _, _, _ = createTestCtx()
	ctx.Env.Domain = "my.test.domain"
	ctx.Flags.Edit = true
	err := preview(ctx, func(path string) error { return nil })
	if assert.NotNil(t, err) {
		assert.Contains(t, err.Error(), "cannot open editor for live theme without theme id")
	}

	ctx, _, _, stdOut, _ := createTestCtx()
	ctx.Env.Domain = "my.test.domain"
	assert.Nil(t, preview(ctx, func(path string) error {
		assert.Equal(t, path, "https://my.test.domain")
		return nil
	}))
	assert.Contains(t, stdOut.String(), "This theme is live so preview is the same as your live shop")

	ctx, _, _, stdOut, _ = createTestCtx()
	ctx.Env.Domain = "my.test.domain"
	err = preview(ctx, func(path string) error {
		assert.Equal(t, path, "https://my.test.domain")
		return fmt.Errorf("fake error")
	})
	assert.Contains(t, stdOut.String(), "This theme is live so preview is the same as your live shop")
	assert.Contains(t, stdOut.String(), "opening")
	assert.Contains(t, err.Error(), "Error opening:")
}

func TestPreviewURL(t *testing.T) {
	testcases := []struct {
		edit         bool
		themeid, url string
	}{
		{edit: true, themeid: "123", url: "https://domain.com/admin/themes/123/editor"},
		{edit: false, themeid: "123", url: "https://domain.com?preview_theme_id=123"},
		{edit: true, url: "https://domain.com/admin/themes//editor"},
		{edit: false, url: "https://domain.com"},
	}

	for _, testcase := range testcases {
		assert.Equal(t, testcase.url, previewURL(testcase.edit, "domain.com", testcase.themeid))
	}
}
