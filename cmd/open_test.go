package cmd

import (
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/Shopify/themekit/kit"
)

type OpenTestSuite struct {
	suite.Suite
}

var (
	testURL string
	wg      sync.WaitGroup
)

func openStubFunc(url string) error {
	fmt.Println("input url: ", url)
	testURL = url
	return nil
}

func (suite *OpenTestSuite) TestOpen() {
	openFunc = openStubFunc

	config, _ := kit.NewConfiguration()
	config.Domain = "my.test.domain"
	config.ThemeID = "123"
	client, _ := kit.NewThemeClient(config)
	wg.Add(4)

	preview(client, []string{}, &wg)
	assert.Equal(suite.T(), "https://my.test.domain?preview_theme_id=123", testURL)

	testURL = ""
	openEdit = true
	preview(client, []string{}, &wg)
	assert.Equal(suite.T(), "https://my.test.domain/admin/themes/123/editor", testURL)

	testURL = ""
	openEdit = false
	config.ThemeID = "live"
	preview(client, []string{}, &wg)
	assert.Equal(suite.T(), "https://my.test.domain?preview_theme_id=", testURL)

	testURL = ""
	openEdit = true
	config.ThemeID = "live"
	preview(client, []string{}, &wg)
	assert.Equal(suite.T(), "", testURL)
}

func TestOpenTestSuite(t *testing.T) {
	suite.Run(t, new(OpenTestSuite))
}
