package phoenix

import (
	"github.com/ryanuber/go-glob"
	"io"
	"io/ioutil"
	"log"
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

func (e EventFilter) Filter(events chan string) chan string {
	filtered := make(chan string)
	go func() {
		for {
			event, more := <-events
			if !more {
				return
			}
			if e.ItemIsValid(event) {
				filtered <- event
			}
		}
	}()
	return filtered
}

func (e EventFilter) ItemIsValid(event string) bool {
	for _, regexp := range e.filters {
		if regexp.MatchString(event) {
			return false
		}
	}
	for _, g := range e.globs {
		if glob.Glob(g, event) {
			return false
		}
	}
	return true
}
