package phoenix

import (
	"github.com/ryanuber/go-glob"
	"io"
	"io/ioutil"
	"log"
	"os"
	re "regexp"
	syn "regexp/syntax"
	"strings"
)

type EventFilter struct {
	filters []*re.Regexp
	globs   []string
}

func NewEventFilter(rawPatterns []string) EventFilter {
	filters := []*re.Regexp{}
	globs := []string{}
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
	return EventFilter{filters: filters, globs: globs}
}

func NewEventFilterFromReaders(readers []io.Reader) EventFilter {
	patterns := []string{}
	for _, reader := range readers {
		data, err := ioutil.ReadAll(reader)
		if err != nil {
			log.Fatal(err)
		}
		otherPatterns := strings.Split(string(data), "\n")
		patterns = append(patterns, otherPatterns...)
	}
	return NewEventFilter(patterns)
}

func NewEventFilterFromFilesCSV(ignores []string) EventFilter {
	files := make([]io.Reader, len(ignores))
	for i, name := range ignores {
		file, err := os.Open(name)
		defer file.Close()
		if err != nil {
			log.Fatal(err, "-", name)
		}
		files[i] = file
	}
	return NewEventFilterFromReaders(files)
}

func (e EventFilter) Filter(events chan string) chan string {
	filtered := make(chan string)
	go func() {
		for {
			event, more := <-events
			if !more {
				return
			}
			if !e.MatchesFilter(event) {
				filtered <- event
			}
		}
	}()
	return filtered
}

func (e EventFilter) MatchesFilter(event string) bool {
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
