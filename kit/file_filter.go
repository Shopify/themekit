package kit

import (
	"io/ioutil"
	"os"
	"regexp"
	"sort"
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
}

var defaultGlobs = []string{}

type fileFilter struct {
	rootDir string
	filters []*regexp.Regexp
	globs   []string
}

func newFileFilter(rootDir string, patterns []string, files []string) (fileFilter, error) {
	filePatterns, err := filesToPatterns(files)
	if err != nil {
		return fileFilter{}, err
	}

	patterns = append(patterns, filePatterns...)

	if !strings.HasSuffix(rootDir, "/") {
		rootDir += "/"
	}

	filters := defaultRegexes
	globs := defaultGlobs
	for _, pattern := range patterns {
		pattern = strings.TrimSpace(pattern)

		// blank lines or comments
		if len(pattern) <= 0 || strings.HasPrefix(pattern, "#") {
			continue
		}

		//full regex
		if strings.HasPrefix(pattern, "/") && strings.HasSuffix(pattern, "/") {
			filters = append(filters, regexp.MustCompile(pattern[1:len(pattern)-1]))
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

		globs = append(globs, rootDir+pattern)
	}

	return fileFilter{
		rootDir: rootDir,
		filters: filters,
		globs:   globs,
	}, nil
}

// filterAssets will filter out compiled assets as well as filter any files that
// match filter patterns.
// It will filter compiled assets by sorting the assets alphabetically and then
// checking that the file after each file does not contain an extra liquid extension.
// For instance if you have the file `app.js` then the file `app.js.liquid`, it
// will filter the first asset (`app.js`) from the slice.
func (e fileFilter) filterAssets(assets []Asset) []Asset {
	filteredAssets := []Asset{}
	sort.Sort(ByAsset(assets))
	for index, asset := range assets {
		if !e.matchesFilter(asset.Key) &&
			(index == len(assets)-1 || assets[index+1].Key != asset.Key+".liquid") {
			filteredAssets = append(filteredAssets, asset)
		}
	}
	return filteredAssets
}

func (e fileFilter) matchesFilter(filename string) bool {
	if len(filename) == 0 || !assetInProject(e.rootDir, filename) {
		return true
	}

	for _, regexp := range e.filters {
		if regexp.MatchString(filename) {
			return true
		}
	}
	for _, pattern := range e.globs {
		if glob.Glob(pattern, filename) || glob.Glob(pattern, e.rootDir+filename) {
			return true
		}
	}

	return false
}

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
		patterns = append(patterns, strings.Split(string(data), "\n")...)
	}
	return patterns, nil
}
