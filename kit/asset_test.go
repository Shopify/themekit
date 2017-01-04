package kit

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type LoadAssetSuite struct {
	suite.Suite
	allocatedFiles []string
}

func (s *LoadAssetSuite) TestIsValid() {
	asset := Asset{Key: "test.txt", Value: "one"}
	assert.Equal(s.T(), true, asset.IsValid())
	asset = Asset{Key: "test.txt", Attachment: "one"}
	assert.Equal(s.T(), true, asset.IsValid())
	asset = Asset{Value: "one"}
	assert.Equal(s.T(), false, asset.IsValid())
	asset = Asset{Key: "test.txt"}
	assert.Equal(s.T(), false, asset.IsValid())
}

func (s *LoadAssetSuite) TestSize() {
	asset := Asset{Value: "one"}
	assert.Equal(s.T(), 3, asset.Size())
	asset = Asset{Attachment: "other"}
	assert.Equal(s.T(), 5, asset.Size())
}

func (s *LoadAssetSuite) TestAssetsSort() {
	input := []Asset{
		{Key: "assets/ajaxify.js.liquid"},
		{Key: "assets/ajaxify.js"},
		{Key: "assets/ajaxify.css"},
		{Key: "assets/ajaxify.css.liquid"},
		{Key: "layouts/customers.liquid"},
	}
	expected := []Asset{
		{Key: "assets/ajaxify.css"},
		{Key: "assets/ajaxify.css.liquid"},
		{Key: "assets/ajaxify.js"},
		{Key: "assets/ajaxify.js.liquid"},
		{Key: "layouts/customers.liquid"},
	}
	sort.Sort(ByAsset(input))
	assert.Equal(s.T(), expected, input)
}

func (s *LoadAssetSuite) TestFindAllFiles() {
	files, err := findAllFiles("../fixtures/project/valid_patterns")
	assert.Equal(s.T(), "Path is not a directory", err.Error())
	files, err = findAllFiles("../fixtures/project")
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), []string{
		clean("../fixtures/project/assets/application.js"),
		clean("../fixtures/project/assets/pixel.png"),
		clean("../fixtures/project/config/settings_data.json"),
		clean("../fixtures/project/config.json"),
		clean("../fixtures/project/invalid_config.yml"),
		clean("../fixtures/project/layout/.gitkeep"),
		clean("../fixtures/project/locales/en.json"),
		clean("../fixtures/project/snippets/snippet.js"),
		clean("../fixtures/project/templates/customers/test.liquid"),
		clean("../fixtures/project/templates/template.liquid"),
		clean("../fixtures/project/valid_config.yml"),
		clean("../fixtures/project/valid_patterns"),
		clean("../fixtures/project/whatever.txt"),
	}, files)
}

func (s *LoadAssetSuite) TestLoadAssetsFromDirectory() {
	assets, err := loadAssetsFromDirectory(clean("../fixtures/project/valid_patterns"), func(path string) bool { return false })
	assert.Equal(s.T(), "Path is not a directory", err.Error())
	assets, err = loadAssetsFromDirectory(clean("../fixtures/project"), func(path string) bool {
		return path != "assets/application.js"
	})
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), []Asset{{
		Key:   "assets/application.js",
		Value: "//this is js\n",
	}}, assets)
}

func (s *LoadAssetSuite) TestLoadAsset() {
	asset, err := loadAsset(clean("../fixtures/project"), clean("assets/application.js"))
	assert.Equal(s.T(), "assets/application.js", asset.Key)
	assert.Equal(s.T(), true, asset.IsValid())
	assert.Equal(s.T(), "//this is js\n", asset.Value)
	assert.Nil(s.T(), err)

	asset, err = loadAsset(clean("../fixtures/project"), "nope.txt")
	assert.NotNil(s.T(), err)

	asset, err = loadAsset(clean("../fixtures/project"), "templates")
	assert.NotNil(s.T(), err)
	assert.Equal(s.T(), "loadAsset: File is a directory", err.Error())

	asset, err = loadAsset(clean("../fixtures/project"), "assets/pixel.png")
	assert.Nil(s.T(), err)
	assert.True(s.T(), len(asset.Attachment) > 0)
	assert.True(s.T(), asset.IsValid())
}

func TestLoadAssetSuite(t *testing.T) {
	suite.Run(t, new(LoadAssetSuite))
}
