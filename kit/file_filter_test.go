package kit

import (
	"path/filepath"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Shopify/themekit/kittest"
)

const validPatterFileContent = `
#plain file names
config/settings.json

#globs
*.png

# regex
/\.(txt|gif|bat)$/
`

func TestNewEventFilter(t *testing.T) {
	kittest.TouchFixtureFile("patterns", validPatterFileContent)
	defer kittest.Cleanup()
	// loads files
	filter, err := newFileFilter(kittest.FixtureProjectPath, []string{}, []string{filepath.Join(kittest.FixtureProjectPath, "patterns")})
	if assert.Nil(t, err) {
		assert.Equal(t, append(defaultRegexes, regexp.MustCompile(`\.(txt|gif|bat)$`)), filter.filters)
		assert.Equal(t, []string{kittest.FixtureProjectPath + "/*config/settings.json", kittest.FixtureProjectPath + "/*.png"}, filter.globs)
	}

	// loads files
	_, err = newFileFilter(kittest.FixtureProjectPath, []string{}, []string{"bad path"})
	assert.NotNil(t, err)

	filter, err = newFileFilter(kittest.FixtureProjectPath, []string{"config/settings.json", "*.png", "/\\.(txt|gif|bat)$/"}, []string{})
	if assert.Nil(t, err) {
		assert.Equal(t, append(defaultRegexes, regexp.MustCompile(`\.(txt|gif|bat)$`)), filter.filters)
		assert.Equal(t, []string{kittest.FixtureProjectPath + "/*config/settings.json", kittest.FixtureProjectPath + "/*.png"}, filter.globs)
	}
}

func TestFilterAssets(t *testing.T) {
	kittest.GenerateProject()
	defer kittest.Cleanup()
	filter, err := newFileFilter(kittest.FixtureProjectPath, []string{".json", "*.txt", "*.gif", "*.ini", "*.bat"}, []string{})
	if assert.Nil(t, err) {
		inputAssets := []Asset{{Key: "templates/foo.json.liquid"}, {Key: "templates/foo.json"}, {Key: "templates/foo.txt"}, {Key: "test.bat"}, {Key: "templates/zubat"}}
		expectedAssets := []Asset{{Key: "templates/foo.json.liquid"}, {Key: "templates/zubat"}}

		assert.Equal(t, expectedAssets, filter.filterAssets(inputAssets))
	}
}

func TestMatchesFilter(t *testing.T) {
	kittest.GenerateProject()
	defer kittest.Cleanup()
	check := func(filter fileFilter, input []string, shouldOutput []string) {
		output := []string{}
		for _, path := range input {
			if !filter.matchesFilter(path) {
				output = append(output, path)
			}
		}
		assert.Equal(t, shouldOutput, output)
	}

	// it filters plain filenames
	filter, err := newFileFilter(kittest.FixtureProjectPath, []string{"build/", "test.txt"}, []string{})
	if assert.Nil(t, err) {
		check(
			filter,
			[]string{kittest.FixtureProjectPath + "/foo/test.txt", "test.txt", "templates/test.txt", "build/hello/world", "build/world", "templates/world", "config/zubat"},
			[]string{"templates/world", "config/zubat"},
		)
	}

	// it filters globs
	filter, err = newFileFilter(kittest.FixtureProjectPath, []string{".json", "*.txt", "*.gif", "*.ini", "*.bat"}, []string{})
	if assert.Nil(t, err) {
		check(
			filter,
			[]string{kittest.FixtureProjectPath + "/config/settings.json", "hello.bat", "build/hello/world.gif", "build/world.txt", "templates/whatever", "foo.ini", "templates/zubat"},
			[]string{"templates/whatever", "templates/zubat"},
		)
	}

	// filters proper regex
	filter, err = newFileFilter(kittest.FixtureProjectPath, []string{`/\.(txt|gif|bat|json|ini)$/`}, []string{})
	if assert.Nil(t, err) {
		check(
			filter,
			[]string{kittest.FixtureProjectPath + "/config/settings.json", "", "hello.bat", "build/hello/world.gif", "build/world.txt", "templates/whatever", "foo.ini", "templates/zubat"},
			[]string{"templates/whatever", "templates/zubat"},
		)
	}

	//check default filters
	filter, err = newFileFilter(kittest.FixtureProjectPath, []string{}, []string{})
	if assert.Nil(t, err) {
		check(
			filter,
			[]string{".git/HEAD", ".DS_Store", "templates/.DS_Store", "config.yml", "templates/products.liquid"},
			[]string{"templates/products.liquid"},
		)
	}

	filter, err = newFileFilter(kittest.FixtureProjectPath, []string{"config/settings_schema.json", "config/settings_data.json", "*.jpg", "*.png"}, []string{})
	if assert.Nil(t, err) {
		assert.Equal(t, true, filter.matchesFilter(""))
	}
}
