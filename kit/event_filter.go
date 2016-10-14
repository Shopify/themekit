package kit

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strings"

	"github.com/ryanuber/go-glob"

	"github.com/Shopify/themekit/theme"
)

const configurationFilename = "config\\.yml"

var defaultRegexes = []*regexp.Regexp{
	regexp.MustCompile(`\.git/*`),
	regexp.MustCompile(`\.DS_Store`),
}

var defaultGlobs = []string{}

type eventFilter struct {
	filters []*regexp.Regexp
	globs   []string
}

func newEventFilter(rootDir string, patterns []string, files []string) eventFilter {
	patterns = append(patterns, filesToPatterns(files)...)

	rootDir += "/"
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
			filters = append(filters, regexp.MustCompile(pattern[1:len(pattern)-2]))
		} else if strings.Contains(pattern, "*") { // globs
			globs = append(globs, rootDir+pattern)
		} else { //plain filename
			globs = append(globs, rootDir+"*"+pattern)
		}
	}
	filters = append(filters, regexp.MustCompile(configurationFilename))
	return eventFilter{filters: filters, globs: globs}
}

func (e eventFilter) filterAssets(assets []theme.Asset) []theme.Asset {
	filteredAssets := []theme.Asset{}
	for _, asset := range assets {
		if !e.matchesFilter(asset.Key) {
			filteredAssets = append(filteredAssets, asset)
		}
	}
	return filteredAssets
}

func (e eventFilter) filter(events chan string) chan string {
	filtered := make(chan string)
	go func() {
		for {
			event, more := <-events
			if !more {
				return
			}
			if len(event) > 0 && !e.matchesFilter(event) {
				filtered <- event
			}
		}
	}()
	return filtered
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
	for _, g := range e.globs {
		if glob.Glob(g, event) {
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

func filesToPatterns(files []string) []string {
	patterns := []string{}
	for _, name := range files {
		file, err := os.Open(name)
		defer file.Close()
		if err != nil {
			Fatal(err)
		}
		var data []byte
		if data, err = ioutil.ReadAll(file); err != nil {
			Fatal(err)
		} else {
			patterns = append(patterns, strings.Split(string(data), "\n")...)
		}
	}
	return patterns
}
