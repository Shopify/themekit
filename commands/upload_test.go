package commands

import (
	"testing"
	"time"

	"github.com/Shopify/themekit/kit"
	"github.com/Shopify/themekit/theme"
	"github.com/stretchr/testify/assert"
)

func TestFullUpload(t *testing.T) {
	assetWithValue := theme.Asset{Key: "layout/layout.liquid", Value: "value1", Attachment: ""}
	assetInSubdir := theme.Asset{Key: "templates/customers/account.liquid", Value: "", Attachment: "attachment"}

	data := []struct {
		local          []theme.Asset
		expectedEvents []kit.AssetEvent
		desc           string
	}{
		{[]theme.Asset{}, []kit.AssetEvent{}, "Empty local and remote, no expected events"},
		{[]theme.Asset{assetWithValue}, []kit.AssetEvent{kit.NewUploadEvent(assetWithValue)}, "Should upload the asset"},
		{[]theme.Asset{assetInSubdir}, []kit.AssetEvent{kit.NewUploadEvent(assetInSubdir)}, "local asset in subdirectory only"},
	}

	for _, d := range data {
		t.Log(d.desc)
		eventCount := 0

		events := make(chan kit.AssetEvent)
		fullUpload(d.local, events)

		select {
		case <-time.After(time.Duration(500) * time.Millisecond):
			t.Logf("Timed out waiting for events.")
			t.Fail()
		case e := <-events:
			if e != nil {
				expectedEvent := d.expectedEvents[eventCount]
				t.Logf("Received %s, expected %s", e.Type(), expectedEvent.Type())
				eventCount++

				assert.Equal(t, expectedEvent.Type(), e.Type(), "Did not get expected event type")
				assert.Equal(t, expectedEvent.Asset().Key, e.Asset().Key)
			}
		}

		assert.Equal(t, len(d.expectedEvents), eventCount, "Did not get the expected number of events!")
	}
}
