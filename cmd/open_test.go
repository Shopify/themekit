package cmd

import (
	"fmt"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Shopify/themekit/kit"
)

func TestOpen(t *testing.T) {
	var wg sync.WaitGroup
	var testURL string

	openFunc = func(url string) error {
		testURL = url
		return nil
	}

	config, _ := kit.NewConfiguration()
	config.Domain = "my.test.domain"
	config.ThemeID = "123"
	client, _ := kit.NewThemeClient(config)
	wg.Add(4)

	err := preview(client, []string{})
	assert.Nil(t, err)
	assert.Equal(t, "https://my.test.domain?preview_theme_id=123", testURL)

	testURL = ""
	openEdit = true
	err = preview(client, []string{})
	assert.Nil(t, err)
	assert.Equal(t, "https://my.test.domain/admin/themes/123/editor", testURL)

	testURL = ""
	openEdit = false
	config.ThemeID = "live"
	err = preview(client, []string{})
	assert.Nil(t, err)
	assert.Equal(t, "https://my.test.domain?preview_theme_id=", testURL)

	testURL = ""
	openEdit = true
	config.ThemeID = "live"
	err = preview(client, []string{})
	assert.True(t, strings.Contains(err.Error(), "cannot open editor for live theme without theme id"))
	assert.Equal(t, "", testURL)

	config.ThemeID = "123"
	openEdit = false
	openFunc = func(string) error { return fmt.Errorf("fake error") }
	err = preview(client, []string{})
	assert.True(t, strings.Contains(err.Error(), "Error opening:"))
}
