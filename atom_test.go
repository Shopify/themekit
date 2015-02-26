package phoenix

import (
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func loadFeedForTesting(t *testing.T) Feed {
	stream, err := os.Open("fixtures/releases.atom")
	assert.Nil(t, err, "Could not load 'fixtures/releases.atom'")
	feed, err := LoadFeed(stream)
	assert.Nil(t, err, "Could not parse 'fixtures/releases.atom'")
	return feed
}

func TestLoadingAtomFeed(t *testing.T) {
	feed := loadFeedForTesting(t)
	assert.Equal(t, "Release notes from Timber", feed.Title)
}

func TestGettingTheLatestEntry(t *testing.T) {
	feed := loadFeedForTesting(t)
	latestEntry := feed.LatestEntry()

	expectedId := "tag:github.com,2008:Repository/17219500/v2.0.2"
	assert.Equal(t, expectedId, latestEntry.Id)

	expectedHref := "/Shopify/Timber/releases/tag/v2.0.2"
	assert.Equal(t, expectedHref, latestEntry.Link.Href)
}
