package kit

import (
	"github.com/Shopify/themekit/theme"
)

// SimpleAssetEvent ... TODO
type SimpleAssetEvent struct {
	asset     theme.Asset
	eventType EventType
}

// Asset ... TODO
func (s SimpleAssetEvent) Asset() theme.Asset {
	return s.asset
}

// Type ... TODO
func (s SimpleAssetEvent) Type() EventType {
	return s.eventType
}

// NewRemovalEvent ... TODO
func NewRemovalEvent(asset theme.Asset) SimpleAssetEvent {
	return SimpleAssetEvent{asset: asset, eventType: Remove}
}

// NewUploadEvent ... TODO
func NewUploadEvent(asset theme.Asset) SimpleAssetEvent {
	return SimpleAssetEvent{asset: asset, eventType: Update}
}
