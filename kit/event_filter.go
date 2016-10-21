package kit

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strings"

	"github.com/ryanuber/go-glob"
)

var defaultRegexes = []*regexp.Regexp{
	regexp.MustCompile(`\.git/*`),
	regexp.MustCompile(`\.DS_Store`),
	regexp.MustCompile(`config.yml`),
}

var defaultGlobs = []string{}

type eventFilter struct {
	rootDir string
	filters []*regexp.Regexp
	globs   []string
}

func newEventFilter(rootDir string, patterns []string, files []string) (eventFilter, error) {
	filePatterns, err := filesToPatterns(files)
	if err != nil {
		return eventFilter{}, err
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

	return eventFilter{
		rootDir: rootDir,
		filters: filters,
		globs:   globs,
	}, nil
}

func (e eventFilter) filterAssets(assets []Asset) []Asset {
	filteredAssets := []Asset{}
	for _, asset := range assets {
		if !e.matchesFilter(asset.Key) {
			filteredAssets = append(filteredAssets, asset)
		}
	}
	return filteredAssets
}

func (e eventFilter) matchesFilter(event string) bool {
	if len(event) == 0 {
		return false
	}
	for _, regexp := range e.filters {
		if regexp.MatchString(event) {
			return true
		}
	}
	for _, pattern := range e.globs {
		if glob.Glob(pattern, event) || glob.Glob(pattern, e.rootDir+event) {
			return true
		}
	}
	return false
}

func (e eventFilter) String() string {
	buffer := bytes.NewBufferString(strings.Join(e.globs, "\n"))
	buffer.WriteString("--- endglobs ---\n")
	for _, rxp := range e.filters {
		buffer.WriteString(fmt.Sprintf("%s\n", rxp))
	}
	buffer.WriteString("-- done --")
	return buffer.String()
}

func filesToPatterns(files []string) ([]string, error) {
	patterns := []string{}
	for _, name := range files {
		file, err := os.Open(name)
		defer file.Close()
		if err != nil {
			return patterns, err
		}
		var data []byte
		if data, err = ioutil.ReadAll(file); err != nil {
			return patterns, err
		}
		patterns = append(patterns, strings.Split(string(data), "\n")...)
	}
	return patterns, nil
}
