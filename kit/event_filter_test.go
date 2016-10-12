package kit

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestEventFilterRejectsEventsThatMatch(t *testing.T) {
	e := newEventFilter([]string{"*.bat", "build/*", "*.ini", "config/settings.json"})

	inputEvents := []string{"path/to/config/settings.json", "hello.bat", "build/hello/world", "build/world", "whatever", "foo.ini", "zubat"}
	expectedEvents := []string{"whatever", "zubat"}
	assertFilter(t, e, inputEvents, expectedEvents)
}

func TestFilterFullRegex(t *testing.T) {
	e := newEventFilter([]string{`/\.(txt|gif|bat)$/`, "config/settings.json", "*.ini"})
	inputEvents := []string{"path/to/config/settings.json", "hello.bat", "build/hello/world.gif", "build/world.txt", "whatever", "foo.ini", "zubat"}
	expectedEvents := []string{"whatever", "zubat"}
	assertFilter(t, e, inputEvents, expectedEvents)
}

func TestFilterRemovesEmptyStrings(t *testing.T) {
	e := newEventFilter([]string{})
	inputEvents := []string{"hello", "", "world"}
	expectedEvents := []string{"hello", "world"}
	assertFilter(t, e, inputEvents, expectedEvents)
}

func TestDefaultFilters(t *testing.T) {
	e := newEventFilter([]string{})
	inputEvents := []string{".git/HEAD", ".DS_Store", "templates/.DS_Store", "config.yml", "templates/products.liquid"}
	expectedEvents := []string{"templates/products.liquid"}
	assertFilter(t, e, inputEvents, expectedEvents)
}

func TestMatchesFilterWithEmptyInputDoesNotCrash(t *testing.T) {
	e := newEventFilter([]string{"config/settings_schema.json", "config/settings_data.json", "*.jpg", "*.png"})
	assert.Equal(t, false, e.matchesFilter(""))
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
	filtered := e.filter(events)

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
