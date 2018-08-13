package file

import (
	"io/ioutil"
	"os"
	"regexp"
	"strings"

	"github.com/ryanuber/go-glob"
)

var defaultRegexes = []*regexp.Regexp{
	regexp.MustCompile(`\.git`),
	regexp.MustCompile(`\.hg`),
	regexp.MustCompile(`\.bzr`),
	regexp.MustCompile(`\.svn`),
	regexp.MustCompile(`_darcs`),
	regexp.MustCompile(`CVS`),
	regexp.MustCompile(`\.sublime-(project|workspace)`),
	regexp.MustCompile(`\.DS_Store`),
	regexp.MustCompile(`\.sass-cache`),
	regexp.MustCompile(`Thumbs\.db`),
	regexp.MustCompile(`desktop\.ini`),
	regexp.MustCompile(`config.yml`),
	regexp.MustCompile(`node_modules`),
}

var defaultGlobs = []string{}

// Filter matches filepaths to a list of patterns
type Filter struct {
	rootDir string
	regexps []*regexp.Regexp
	globs   []string
}

// NewFilter will create a new file path filter
func NewFilter(rootDir string, patterns []string, files []string) (Filter, error) {
	filePatterns, err := filesToPatterns(files)
	if err != nil {
		return Filter{}, err
	}

	if !strings.HasSuffix(rootDir, "/") {
		rootDir += "/"
	}

	regexps, globs := patternsToRegexpsAndGlobs(append(patterns, filePatterns...))

	return Filter{
		rootDir: rootDir,
		regexps: regexps,
		globs:   globs,
	}, nil
}

// Match will return true if the file path has matched a pattern in this filter
func (f Filter) Match(path string) bool {
	if len(path) == 0 || !pathInProject(f.rootDir, path) {
		return true
	}

	for _, regexp := range f.regexps {
		if regexp.MatchString(path) {
			return true
		}
	}

	for _, pattern := range f.globs {
		if glob.Glob(pattern, path) {
			return true
		}
	}

	return false
}

// filesToPatterns will load up external files and scrape patterns from them
func filesToPatterns(files []string) ([]string, error) {
	patterns := []string{}
	for _, name := range files {
		file, err := os.Open(name)
		if err != nil {
			return patterns, err
		}
		defer file.Close()
		var data []byte
		if data, err = ioutil.ReadAll(file); err != nil {
			return patterns, err
		}

		for _, line := range strings.Split(string(data), "\n") {
			line = strings.TrimSuffix(line, "\r") // remove windows carraige return
			if len(line) > 0 && !strings.HasPrefix(line, "#") {
				patterns = append(patterns, line)
			}
		}
	}
	return patterns, nil
}

// patternsToFiltersAndGlobs will take in string patterns and convert them to either
// regex patters or glob patterns so that they are handled in an expected manner.
func patternsToRegexpsAndGlobs(patterns []string) ([]*regexp.Regexp, []string) {
	regexps := defaultRegexes
	globs := defaultGlobs

	for _, pattern := range patterns {
		pattern = strings.TrimSpace(pattern)

		//full regex
		if strings.HasPrefix(pattern, "/") && strings.HasSuffix(pattern, "/") {
			regexps = append(regexps, regexp.MustCompile(pattern[1:len(pattern)-1]))
			continue
		}

		// if specifying a directory match everything below it
		if strings.HasSuffix(pattern, "/") {
			pattern += "*"
		}

		// The pattern will be scoped to root directory so it should match anything
		// within that space
		if !strings.HasPrefix(pattern, "*") {
			pattern = "*" + pattern
		}

		globs = append(globs, pattern)
	}

	return regexps, globs
}
