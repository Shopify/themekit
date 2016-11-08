package kit

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

const (
	rootDir           = "./root/dir/"
	ignoreFixturePath = "../fixtures/project/valid_patterns"
)

type EventFilterTestSuite struct {
	suite.Suite
}

func (suite *EventFilterTestSuite) TestNewEventFilter() {
	// loads files
	filter, err := newEventFilter(rootDir, []string{}, []string{ignoreFixturePath})
	if assert.Nil(suite.T(), err) {
		assert.Equal(suite.T(), append(defaultRegexes, regexp.MustCompile(`\.(txt|gif|bat)$`)), filter.filters)
		assert.Equal(suite.T(), []string{"./root/dir/*config/settings.json", "./root/dir/*.png"}, filter.globs)
	}

	// loads files
	_, err = newEventFilter(rootDir, []string{}, []string{"bad path"})
	assert.NotNil(suite.T(), err)

	filter, err = newEventFilter(rootDir, []string{"config/settings.json", "*.png", "/\\.(txt|gif|bat)$/"}, []string{})
	if assert.Nil(suite.T(), err) {
		assert.Equal(suite.T(), append(defaultRegexes, regexp.MustCompile(`\.(txt|gif|bat)$`)), filter.filters)
		assert.Equal(suite.T(), []string{"./root/dir/*config/settings.json", "./root/dir/*.png"}, filter.globs)
	}
}

func (suite *EventFilterTestSuite) TestFilterAssets() {
	filter, err := newEventFilter(rootDir, []string{".json", "*.txt", "*.gif", "*.ini", "*.bat"}, []string{})
	if assert.Nil(suite.T(), err) {
		inputAssets := []Asset{{Key: "test/foo"}, {Key: "foo.txt"}, {Key: "test.bat"}, {Key: "zubat"}}
		expectedAssets := []Asset{{Key: "test/foo"}, {Key: "zubat"}}

		assert.Equal(suite.T(), expectedAssets, filter.filterAssets(inputAssets))
	}
}

func (suite *EventFilterTestSuite) TestMatchesFilter() {
	check := func(filter eventFilter, input []string, shouldOutput []string) {
		output := []string{}
		for _, path := range input {
			if !filter.matchesFilter(path) {
				output = append(output, path)
			}
		}
		assert.Equal(suite.T(), shouldOutput, output)
	}

	// it filters plain filenames
	filter, err := newEventFilter(rootDir, []string{"build/", "test.txt"}, []string{})
	if assert.Nil(suite.T(), err) {
		check(
			filter,
			[]string{rootDir + "/foo/test.txt", "test.txt", "test/test.txt", "build/hello/world", "build/world", "test/build/world", "zubat"},
			[]string{"zubat"},
		)
	}

	// it filters globs
	filter, err = newEventFilter(rootDir, []string{".json", "*.txt", "*.gif", "*.ini", "*.bat"}, []string{})
	if assert.Nil(suite.T(), err) {
		check(
			filter,
			[]string{rootDir + "config/settings.json", "hello.bat", "build/hello/world.gif", "build/world.txt", "whatever", "foo.ini", "zubat"},
			[]string{"whatever", "zubat"},
		)
	}

	// filters proper regex
	filter, err = newEventFilter(rootDir, []string{`/\.(txt|gif|bat|json|ini)$/`}, []string{})
	if assert.Nil(suite.T(), err) {
		check(
			filter,
			[]string{rootDir + "config/settings.json", "hello.bat", "build/hello/world.gif", "build/world.txt", "whatever", "foo.ini", "zubat"},
			[]string{"whatever", "zubat"},
		)
	}

	//check default filters
	filter, err = newEventFilter(rootDir, []string{}, []string{})
	if assert.Nil(suite.T(), err) {
		check(
			filter,
			[]string{".git/HEAD", ".DS_Store", "templates/.DS_Store", "templates/products.liquid"},
			[]string{"templates/products.liquid"},
		)
	}

	filter, err = newEventFilter(rootDir, []string{"config/settings_schema.json", "config/settings_data.json", "*.jpg", "*.png"}, []string{})
	if assert.Nil(suite.T(), err) {
		assert.Equal(suite.T(), false, filter.matchesFilter(""))
	}
}

func TestEventFilterTestSuite(t *testing.T) {
	suite.Run(t, new(EventFilterTestSuite))
}
