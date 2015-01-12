package phoenix

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"io"
	"testing"
	"time"
)

func TestEventFilterRejectsEventsThatMatch(t *testing.T) {
	eventFilter := NewEventFilter([]string{"foo", "baa*"})

	inputEvents := []string{"hello", "foo", "gofoo", "baarber", "barber", "goodbye"}
	expectedEvents := []string{"hello", "goodbye"}
	assertFilter(t, eventFilter, inputEvents, expectedEvents)
}

func TestEventFilterTurnsInvalidRegexpsIntoGlobs(t *testing.T) {
	eventFilter := NewEventFilter([]string{"*.bat", "build/*", "*.ini"})

	inputEvents := []string{"hello.bat", "build/hello/world", "build/world", "whatever", "foo.ini", "zubat"}
	expectedEvents := []string{"whatever", "zubat"}
	assertFilter(t, eventFilter, inputEvents, expectedEvents)
}

func TestBuildingEventFiltersFromMultipleReaders(t *testing.T) {
	readers := []io.Reader{
		bytes.NewReader([]byte("*.bat\nbuild/")),
		bytes.NewReader([]byte("foo\nbar")),
	}
	eventFilter := NewEventFilterFromReaders(readers)
	inputEvents := []string{
		"program.bat", "build/dist/program", "item.liquid", "gofoo", "gobar", "listing", "programbat", "config.yml",
	}
	expectedResults := []string{"item.liquid", "listing", "programbat"}
	assertFilter(t, eventFilter, inputEvents, expectedResults)
}

func TestFilterRemovesEmptyStrings(t *testing.T) {
	eventFilter := NewEventFilterFromReaders([]io.Reader{})
	inputEvents := []string{"hello", "", "world"}
	expectedEvents := []string{"hello", "world"}
	assertFilter(t, eventFilter, inputEvents, expectedEvents)
}

func nextValue(channel chan string) string {
	select {
	case result := <-channel:
		return result
	case <-time.After(10 * time.Millisecond):
		return ""
	}
}

func assertFilter(t *testing.T, eventFilter EventFilter, inputs []string, expectedResults []string) {
	events := make(chan string)
	filtered := eventFilter.Filter(events)

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
