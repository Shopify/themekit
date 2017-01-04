package kit

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

var (
	rootDir           = "./root/dir/"
	ignoreFixturePath = "../fixtures/project/valid_patterns"
)

type EventFilterTestSuite struct {
	suite.Suite
}

func (suite *EventFilterTestSuite) TestNewEventFilter() {
	// loads files
	filter, err := newFileFilter(rootDir, []string{}, []string{ignoreFixturePath})
	if assert.Nil(suite.T(), err) {
		assert.Equal(suite.T(), append(defaultRegexes, regexp.MustCompile(`\.(txt|gif|bat)$`)), filter.filters)
		assert.Equal(suite.T(), []string{"./root/dir/*config/settings.json", "./root/dir/*.png"}, filter.globs)
	}

	// loads files
	_, err = newFileFilter(rootDir, []string{}, []string{"bad path"})
	assert.NotNil(suite.T(), err)

	filter, err = newFileFilter(rootDir, []string{"config/settings.json", "*.png", "/\\.(txt|gif|bat)$/"}, []string{})
	if assert.Nil(suite.T(), err) {
		assert.Equal(suite.T(), append(defaultRegexes, regexp.MustCompile(`\.(txt|gif|bat)$`)), filter.filters)
		assert.Equal(suite.T(), []string{"./root/dir/*config/settings.json", "./root/dir/*.png"}, filter.globs)
	}
}

func (suite *EventFilterTestSuite) TestFilterAssets() {
	filter, err := newFileFilter(rootDir, []string{".json", "*.txt", "*.gif", "*.ini", "*.bat"}, []string{})
	if assert.Nil(suite.T(), err) {
		inputAssets := []Asset{{Key: "templates/foo.json.liquid"}, {Key: "templates/foo.json"}, {Key: "templates/foo.txt"}, {Key: "test.bat"}, {Key: "templates/zubat"}}
		expectedAssets := []Asset{{Key: "templates/foo.json.liquid"}, {Key: "templates/zubat"}}

		assert.Equal(suite.T(), expectedAssets, filter.filterAssets(inputAssets))
	}
}

func (suite *EventFilterTestSuite) TestMatchesFilter() {
	check := func(filter fileFilter, input []string, shouldOutput []string) {
		output := []string{}
		for _, path := range input {
			if !filter.matchesFilter(path) {
				output = append(output, path)
			}
		}
		assert.Equal(suite.T(), shouldOutput, output)
	}

	// it filters plain filenames
	filter, err := newFileFilter(rootDir, []string{"build/", "test.txt"}, []string{})
	if assert.Nil(suite.T(), err) {
		check(
			filter,
			[]string{rootDir + "/foo/test.txt", "test.txt", "templates/test.txt", "build/hello/world", "build/world", "templates/world", "config/zubat"},
			[]string{"templates/world", "config/zubat"},
		)
	}

	// it filters globs
	filter, err = newFileFilter(rootDir, []string{".json", "*.txt", "*.gif", "*.ini", "*.bat"}, []string{})
	if assert.Nil(suite.T(), err) {
		check(
			filter,
			[]string{rootDir + "config/settings.json", "hello.bat", "build/hello/world.gif", "build/world.txt", "templates/whatever", "foo.ini", "templates/zubat"},
			[]string{"templates/whatever", "templates/zubat"},
		)
	}

	// filters proper regex
	filter, err = newFileFilter(rootDir, []string{`/\.(txt|gif|bat|json|ini)$/`}, []string{})
	if assert.Nil(suite.T(), err) {
		check(
			filter,
			[]string{rootDir + "config/settings.json", "", "hello.bat", "build/hello/world.gif", "build/world.txt", "templates/whatever", "foo.ini", "templates/zubat"},
			[]string{"templates/whatever", "templates/zubat"},
		)
	}

	//check default filters
	filter, err = newFileFilter(rootDir, []string{}, []string{})
	if assert.Nil(suite.T(), err) {
		check(
			filter,
			[]string{".git/HEAD", ".DS_Store", "templates/.DS_Store", "config.yml", "templates/products.liquid"},
			[]string{"templates/products.liquid"},
		)
	}

	filter, err = newFileFilter(rootDir, []string{"config/settings_schema.json", "config/settings_data.json", "*.jpg", "*.png"}, []string{})
	if assert.Nil(suite.T(), err) {
		assert.Equal(suite.T(), true, filter.matchesFilter(""))
	}
}

func TestEventFilterTestSuite(t *testing.T) {
	suite.Run(t, new(EventFilterTestSuite))
}
