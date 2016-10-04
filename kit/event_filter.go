package kit

import (
	"bytes"
	"fmt"
	"io"
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

// EventFilter ... TODO
type EventFilter struct {
	filters []*re.Regexp
	globs   []string
}

// NewEventFilter ... TODO
func NewEventFilter(rawPatterns []string) EventFilter {
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
	return EventFilter{filters: filters, globs: globs}
}

// NewEventFilterFromReaders ... TODO
func NewEventFilterFromReaders(readers []io.Reader) EventFilter {
	patterns := []string{}
	for _, reader := range readers {
		data, err := ioutil.ReadAll(reader)
		if err != nil {
			NotifyError(err)
		}
		otherPatterns := strings.Split(string(data), "\n")
		patterns = append(patterns, otherPatterns...)
	}
	return NewEventFilter(patterns)
}

// NewEventFilterFromIgnoreFiles ... TODO
func NewEventFilterFromIgnoreFiles(ignores []string) EventFilter {
	files := filenamesToReaders(ignores)
	return NewEventFilterFromReaders(files)
}

// NewEventFilterFromPatternsAndFiles ... TODO
func NewEventFilterFromPatternsAndFiles(patterns []string, files []string) EventFilter {
	readers := filenamesToReaders(files)
	allReaders := make([]io.Reader, len(readers)+len(patterns))
	pos := 0
	for i := 0; i < len(readers); i++ {
		allReaders[pos] = readers[i]
		pos++
	}
	for i := 0; i < len(patterns); i++ {
		allReaders[pos] = strings.NewReader(patterns[i])
		pos++
	}
	return NewEventFilterFromReaders(allReaders)
}

func (e EventFilter) FilterAssets(assets []theme.Asset) []theme.Asset {
	filteredAssets := []theme.Asset{}
	for _, asset := range assets {
		if !e.MatchesFilter(asset.Key) {
			filteredAssets = append(filteredAssets, asset)
		}
	}
	return filteredAssets
}

// Filter ... TODO
func (e EventFilter) Filter(events chan string) chan string {
	filtered := make(chan string)
	go func() {
		for {
			event, more := <-events
			if !more {
				return
			}
			if len(event) > 0 && !e.MatchesFilter(event) {
				filtered <- event
			}
		}
	}()
	return filtered
}

// MatchesFilter ... TODO
func (e EventFilter) MatchesFilter(event string) bool {
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

func (e EventFilter) String() string {
	buffer := bytes.NewBufferString(strings.Join(e.globs, "\n"))
	buffer.WriteString("--- endglobs ---\n")
	for _, rxp := range e.filters {
		buffer.WriteString(fmt.Sprintf("%s\n", rxp))
	}
	buffer.WriteString("-- done --")
	return buffer.String()
}

func filenamesToReaders(ignores []string) []io.Reader {
	files := make([]io.Reader, len(ignores))
	for i, name := range ignores {
		file, err := os.Open(name)
		defer file.Close()
		if err != nil {
			NotifyError(err)
		}
		files[i] = file
	}
	return files
}
