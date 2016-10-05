package kit

import (
	"github.com/Shopify/themekit/theme"
)

type simpleAssetEvent struct {
	asset     theme.Asset
	eventType EventType
}

// Asset ... TODO
func (s simpleAssetEvent) Asset() theme.Asset {
	return s.asset
}

// Type ... TODO
func (s simpleAssetEvent) Type() EventType {
	return s.eventType
}

// NewRemovalEvent ... TODO
func NewRemovalEvent(asset theme.Asset) AssetEvent {
	return simpleAssetEvent{asset: asset, eventType: Remove}
}

// NewUploadEvent ... TODO
func NewUploadEvent(asset theme.Asset) AssetEvent {
	return simpleAssetEvent{asset: asset, eventType: Update}
}
