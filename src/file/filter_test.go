package file

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewFilter(t *testing.T) {
	expected := Filter{
		rootDir: "/tmp/",
		regexps: defaultRegexes,
		globs:   defaultGlobs,
	}
	actual, err := NewFilter("/tmp", []string{}, []string{})
	assert.Nil(t, err)
	assert.Equal(t, expected, actual)

	_, err = NewFilter("/tmp", []string{}, []string{"does not exists"})
	assert.NotNil(t, err)
}

func TestFilter_Match(t *testing.T) {
	testcases := []struct {
		regexp, glob, input string
		matches             bool
	}{
		{glob: "*test.txt", input: "templates/test.txt", matches: true},
		{glob: "*test.txt", input: "templates/foo/test.txt", matches: true},
		{glob: "*test.txt", input: "/tmp/templates/foo/test.txt", matches: true},
		{glob: "*build/*", input: "templates/build/hello/world", matches: true},
		{glob: "*.json", input: "templates/settings.json", matches: true},
		{glob: "*.gif", input: "templates/world.gif", matches: true},
		{glob: "*.gif", input: "templates/worldgifno", matches: false},
		{regexp: `\.bat`, input: "templates/hello.bat", matches: true},
		{regexp: `\.bat`, input: "templates/hellobatno", matches: false},
		{regexp: `\.bat`, input: "templates/hello.css", matches: false},
		{glob: "*test.txt", input: "/not/in/project/test.txt", matches: true},
		{glob: "*test.txt", input: "test.txt", matches: true},
		{input: "", matches: true},
	}

	for _, testcase := range testcases {
		filter := Filter{rootDir: "/tmp"}
		if testcase.regexp != "" {
			filter.regexps = []*regexp.Regexp{regexp.MustCompile(testcase.regexp)}
		}
		if testcase.glob != "" {
			filter.globs = []string{testcase.glob}
		}
		if testcase.matches {
			assert.True(t, filter.Match(testcase.input), testcase.input)
		} else {
			assert.False(t, filter.Match(testcase.input), testcase.input)
		}
	}
}

func TestFilesToPatterns(t *testing.T) {
	patterns, err := filesToPatterns([]string{"_testdata/ignores_file"})
	assert.Nil(t, err)
	assert.Equal(t, patterns, []string{"config/settings.json", "*.png", `/\.(txt|gif|bat)$/`})

	_, err = filesToPatterns([]string{"does not exist"})
	assert.NotNil(t, err)
}

func TestPatternsToRegexpsAndGlobs(t *testing.T) {
	testcases := []struct {
		pattern string
		glob    string
		regex   *regexp.Regexp
	}{
		{pattern: "config/settings.json", glob: "*config/settings.json"},
		{pattern: "config/", glob: "*config/*"},
		{pattern: "*.png", glob: "*.png"},
		{pattern: `/\.(txt|gif|bat)$/`, regex: regexp.MustCompile(`\.(txt|gif|bat)$`)},
	}

	for _, testcase := range testcases {
		regexps, globs := patternsToRegexpsAndGlobs([]string{testcase.pattern})
		if testcase.regex != nil {
			assert.Equal(t, append(defaultRegexes, testcase.regex), regexps)
		}
		if testcase.glob != "" {
			assert.Equal(t, append(defaultGlobs, testcase.glob), globs)
		}
	}
}
