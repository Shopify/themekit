package commands

import (
	"testing"
	"time"

	"github.com/Shopify/themekit"
	"github.com/Shopify/themekit/theme"
	"github.com/stretchr/testify/assert"
)

func TestFullReplace(t *testing.T) {
	assetWithValue := theme.Asset{Key: "layout/layout.liquid", Value: "value1", Attachment: ""}
	assetWithAttachment := theme.Asset{Key: "layout/layout.liquid", Value: "", Attachment: "attachment"}
	assetInSubdir := theme.Asset{Key: "templates/customers/account.liquid", Value: "", Attachment: "attachment"}

	data := []struct {
		local          []theme.Asset
		remote         []theme.Asset
		expectedEvents []themekit.AssetEvent
		desc           string
	}{
		{[]theme.Asset{}, []theme.Asset{}, []themekit.AssetEvent{}, "Empty local and remote, no expected events"},
		{[]theme.Asset{assetWithValue}, []theme.Asset{assetWithValue}, []themekit.AssetEvent{themekit.NewUploadEvent(assetWithValue)}, "Both local and remote are the same"},
		{[]theme.Asset{assetWithValue}, []theme.Asset{}, []themekit.AssetEvent{themekit.NewUploadEvent(assetWithValue)}, "Only local asset exists"},
		{[]theme.Asset{}, []theme.Asset{assetWithValue}, []themekit.AssetEvent{themekit.NewRemovalEvent(assetWithValue)}, "Asset exists only remotely"},
		{[]theme.Asset{assetWithAttachment}, []theme.Asset{assetWithValue}, []themekit.AssetEvent{themekit.NewUploadEvent(assetWithValue)}, "Local asset has attachment, remote value"},
		{[]theme.Asset{assetWithValue}, []theme.Asset{assetWithAttachment}, []themekit.AssetEvent{themekit.NewUploadEvent(assetWithValue)}, "Local asset has value, remote has attachment"},
		{[]theme.Asset{assetInSubdir}, []theme.Asset{}, []themekit.AssetEvent{themekit.NewUploadEvent(assetInSubdir)}, "local asset in subdirectory only"},
		{[]theme.Asset{}, []theme.Asset{assetInSubdir}, []themekit.AssetEvent{themekit.NewRemovalEvent(assetInSubdir)}, "remote asset in subdirectory only"},
		{[]theme.Asset{assetInSubdir}, []theme.Asset{assetInSubdir}, []themekit.AssetEvent{themekit.NewUploadEvent(assetInSubdir)}, "local asset in subdirectory both local and remote"},
	}

	for _, d := range data {
		t.Log(d.desc)
		eventCount := 0

		events := make(chan themekit.AssetEvent)
		fullReplace(d.remote, d.local, events)

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
