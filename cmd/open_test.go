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

	ctx, _, _, stdOut, _ := createTestCtx()
	ctx.Env.Domain = "my.test.domain"
	ctx.Env.ThemeID = "123"
	err := preview(ctx, func(path string) error {
		assert.Equal(t, path, "https://my.test.domain?preview_theme_id=123")
		return fmt.Errorf("fake error")
	})
	assert.Contains(t, stdOut.String(), "opening")
	assert.Contains(t, err.Error(), "Error opening:")
}
