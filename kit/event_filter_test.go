package kit

import (
	"bytes"
	"io"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestEventFilterRejectsEventsThatMatch(t *testing.T) {
	e := newEventFilter([]string{"foo", "baa*"})

	inputEvents := []string{"hello", "foo", "gofoo", "baarber", "barber", "goodbye"}
	expectedEvents := []string{"hello", "goodbye"}
	assertFilter(t, e, inputEvents, expectedEvents)
}

func TestEventFilterTurnsInvalidRegexpsIntoGlobs(t *testing.T) {
	e := newEventFilter([]string{"*.bat", "build/*", "*.ini"})

	inputEvents := []string{"hello.bat", "build/hello/world", "build/world", "whatever", "foo.ini", "zubat"}
	expectedEvents := []string{"whatever", "zubat"}
	assertFilter(t, e, inputEvents, expectedEvents)
}

func TestBuildingEventFiltersFromMultipleReaders(t *testing.T) {
	readers := []io.Reader{
		bytes.NewReader([]byte("*.bat\nbuild/")),
		bytes.NewReader([]byte("foo\nbar")),
	}
	e := newEventFilterFromReaders(readers)
	inputEvents := []string{
		"program.bat", "build/dist/program", "item.liquid", "gofoo", "gobar", "listing", "programbat", "config.yml",
	}
	expectedResults := []string{"item.liquid", "listing", "programbat"}
	assertFilter(t, e, inputEvents, expectedResults)
}

func TestFilterRemovesEmptyStrings(t *testing.T) {
	e := newEventFilterFromReaders([]io.Reader{})
	inputEvents := []string{"hello", "", "world"}
	expectedEvents := []string{"hello", "world"}
	assertFilter(t, e, inputEvents, expectedEvents)
}

func TestDefaultFilters(t *testing.T) {
	e := newEventFilterFromReaders([]io.Reader{})
	inputEvents := []string{".git/HEAD", ".DS_Store", "config.yml", "templates/products.liquid"}
	expectedEvents := []string{"templates/products.liquid"}
	assertFilter(t, e, inputEvents, expectedEvents)
}

func TestMatchesFilterWithEmptyInputDoesNotCrash(t *testing.T) {
	e := newEventFilter([]string{"config/settings_schema.json", "config/settings_data.json", "*.jpg", "*.png"})
	// Shouldn't crash
	e.MatchesFilter("")
}

func nextValue(channel chan string) string {
	select {
	case result := <-channel:
		return result
	case <-time.After(10 * time.Millisecond):
		return ""
	}
}

func assertFilter(t *testing.T, e eventFilter, inputs []string, expectedResults []string) {
	events := make(chan string)
	filtered := e.Filter(events)

	go func() {
		for _, event := range inputs {
			events <- event
		}
		close(events)
	}()

	for _, expected := range expectedResults {
		assert.Equal(t, expected, nextValue(filtered))
	}
}
