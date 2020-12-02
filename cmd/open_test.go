package cmd

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOpen(t *testing.T) {
	r := func(path string) error { panic("should not have been called") }
	rw := func(path, with string) error { panic("should not have been called") }

	ctx, _, _, _, _ := createTestCtx()
	ctx.Env.Domain = "my.test.domain"
	ctx.Env.ThemeID = "123"
	assert.Nil(t, preview(ctx, func(path string) error {
		assert.Equal(t, path, "https://my.test.domain?preview_theme_id=123")
		return nil
	}, rw))

	ctx, _, _, _, _ = createTestCtx()
	ctx.Env.Domain = "my.test.domain"
	ctx.Env.ThemeID = "123"
	ctx.Flags.HidePreviewBar = true
	assert.Nil(t, preview(ctx, func(path string) error {
		assert.Equal(t, path, "https://my.test.domain?preview_theme_id=123&pb=0")
		return nil
	}, rw))

	ctx, _, _, _, _ = createTestCtx()
	ctx.Env.Domain = "my.test.domain"
	ctx.Env.ThemeID = "123"
	ctx.Flags.Edit = true
	assert.Nil(t, preview(ctx, func(path string) error {
		assert.Equal(t, path, "https://my.test.domain/admin/themes/123/editor")
		return nil
	}, rw))

	ctx, _, _, stdOut, _ := createTestCtx()
	ctx.Env.Domain = "my.test.domain"
	ctx.Env.ThemeID = "123"
	err := preview(ctx, func(path string) error {
		assert.Equal(t, path, "https://my.test.domain?preview_theme_id=123")
		return fmt.Errorf("fake error")
	}, rw)
	assert.Contains(t, stdOut.String(), "opening")
	assert.Contains(t, err.Error(), "Error opening:")

	ctx, _, _, stdOut, _ = createTestCtx()
	ctx.Env.Domain = "my.test.domain"
	ctx.Env.ThemeID = "123"
	ctx.Flags.With = "chrome"
	err = preview(ctx, r, func(path, with string) error {
		assert.Equal(t, path, "https://my.test.domain?preview_theme_id=123")
		assert.Equal(t, with, "chrome")
		return fmt.Errorf("fake error")
	})
	assert.Contains(t, stdOut.String(), "opening")
	assert.Contains(t, err.Error(), "Error opening:")
}
