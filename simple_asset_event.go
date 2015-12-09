package themekit

import (
	"github.com/Shopify/themekit/theme"
)

type SimpleAssetEvent struct {
	asset     theme.Asset
	eventType EventType
}

func (s SimpleAssetEvent) Asset() theme.Asset {
	return s.asset
}

func (s SimpleAssetEvent) Type() EventType {
	return s.eventType
}

func NewRemovalEvent(asset theme.Asset) SimpleAssetEvent {
	return SimpleAssetEvent{asset: asset, eventType: Remove}
}

func NewUploadEvent(asset theme.Asset) SimpleAssetEvent {
	return SimpleAssetEvent{asset: asset, eventType: Update}
}
