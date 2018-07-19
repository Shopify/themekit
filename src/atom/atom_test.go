package atom

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

const testAtom = `<?xml version="1.0" encoding="UTF-8"?>
<feed xmlns="http://www.w3.org/2005/Atom" xmlns:media="http://search.yahoo.com/mrss/" xml:lang="en-US">
  <title>Release notes from Timber</title>
  <entry>
    <id>tag:github.com,2008:Repository/17219500/v2.0.2</id>
    <link rel="alternate" type="text/html" href="/Shopify/Timber/releases/tag/v2.0.2"/>
    <title>v2.0.2</title>
  </entry>
</feed>
`

func loadFeedForTesting(t *testing.T) Feed {
	feed, err := LoadFeed(strings.NewReader(testAtom))
	assert.Nil(t, err, "Could not parse 'releases.atom'")
	return feed
}

func TestLoadingAtomFeed(t *testing.T) {
	feed := loadFeedForTesting(t)
	assert.Equal(t, "Release notes from Timber", feed.Title)
}

func TestGettingTheLatestEntry(t *testing.T) {
	feed := loadFeedForTesting(t)
	latestEntry := feed.LatestEntry()

	expectedID := "tag:github.com,2008:Repository/17219500/v2.0.2"
	assert.Equal(t, expectedID, latestEntry.ID)

	expectedHref := "/Shopify/Timber/releases/tag/v2.0.2"
	assert.Equal(t, expectedHref, latestEntry.Link.Href)
}
