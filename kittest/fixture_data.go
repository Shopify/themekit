package kittest

import (
	"strings"

	"github.com/Shopify/themekit/cmd/atom"
)

var (
	releaseAtom = `<?xml version="1.0" encoding="UTF-8"?>
<feed xmlns="http://www.w3.org/2005/Atom" xmlns:media="http://search.yahoo.com/mrss/" xml:lang="en-US">
  <entry> <title>v2.0.2</title> </entry>
  <entry><title>v2.0.1</title></entry>
  <entry><title>v2.0.0</title></entry>
  <entry><title>v1.3.2</title></entry>
  <entry><title>v1.3.1</title></entry>
  <entry><title>v1.3.0</title></entry>
  <entry><title>v1.2.1</title></entry>
  <entry><title>v1.2.0</title></entry>
  <entry><title>v1.1.3</title></entry>
  <entry><title>v1.1.2</title></entry>
  <entry><title>v1.1.1</title></entry>
  <entry><title>v1.1.0</title></entry>
  <entry><title>v1.0.0</title></entry>
</feed>`
	// ReleaseAtom is a atom release for testing release selection
	ReleaseAtom, _ = atom.LoadFeed(strings.NewReader(releaseAtom))
	themesResponse = `{
  "themes":[
    {
      "theme": {
        "name": "timberland",
        "source": "https://githubz.com/shopify/timberlands",
        "role": "unpublished",
        "previewable": true
      }
    }
  ]
}`
	themeResponse = `{
  "theme": {
    "name": "timberland",
    "source": "https://githubz.com/shopify/timberlands",
    "role": "unpublished",
    "previewable": true
  }
}`
	themeErrorResponse = `{ "errors":{ "src":[ "is empty" ] } }`
	assetsReponse      = `{
  "assets": [
    { "key": "assets/hello.txt", "value": "Hello World", "updated_at":"2012-07-06T02:04:21-11:00" },
    { "key": "assets/goodbye.txt", "value": "Goodbye", "updated_at":"2012-07-06T02:04:21-11:00" }
  ]
} `
	assetResponse = `{
  "asset": {
    "key": "assets/hello.txt",
    "value": "hello world",
		"updated_at":"2012-07-06T02:04:21-11:00",
    "warnings": []
  }
}`
	themekitUpdateFeed = `[{"version":"0.4.4"},{"version":"0.4.7", "platforms": [{}]},{"version":"0.4.6"},{"version":"0.4.5"}]`
)
