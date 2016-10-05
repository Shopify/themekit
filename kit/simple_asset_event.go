package kit

import (
	"github.com/Shopify/themekit/theme"
)

type simpleAssetEvent struct {
	asset     theme.Asset
	eventType EventType
}

func (s simpleAssetEvent) Asset() theme.Asset {
	return s.asset
}

func (s simpleAssetEvent) Type() EventType {
	return s.eventType
}

// NewRemovalEvent will create a simple asset removal event for the theme client
// to process
func NewRemovalEvent(asset theme.Asset) AssetEvent {
	return simpleAssetEvent{asset: asset, eventType: Remove}
}

// NewUploadEvent will create a simple asset update event for the theme client
// to process
func NewUploadEvent(asset theme.Asset) AssetEvent {
	return simpleAssetEvent{asset: asset, eventType: Update}
}
