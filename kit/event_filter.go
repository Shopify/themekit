package kit

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	re "regexp"
	syn "regexp/syntax"
	"strings"

	"github.com/ryanuber/go-glob"

	"github.com/Shopify/themekit/theme"
)

const configurationFilename = "config\\.yml"

var defaultRegexes = []*re.Regexp{
	re.MustCompile(`\.git/*`),
	re.MustCompile(`\.DS_Store`),
}

var defaultGlobs = []string{}

type eventFilter struct {
	filters []*re.Regexp
	globs   []string
}

func newEventFilter(rawPatterns []string) eventFilter {
	filters := defaultRegexes
	globs := defaultGlobs
	for _, pat := range rawPatterns {
		if len(pat) <= 0 {
			continue
		}
		regex, err := syn.Parse(pat, syn.POSIX)
		if err != nil {
			globs = append(globs, pat)
		} else {
			filters = append(filters, re.MustCompile(regex.String()))
		}
	}
	filters = append(filters, re.MustCompile(configurationFilename))
	return eventFilter{filters: filters, globs: globs}
}

func newEventFilterFromPatternsAndFiles(patterns []string, files []string) eventFilter {
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
	return newEventFilter(patterns)
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
