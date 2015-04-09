package themekit

type SimpleAssetEvent struct {
	asset     Asset
	eventType EventType
}

func (s SimpleAssetEvent) Asset() Asset {
	return s.asset
}

func (s SimpleAssetEvent) Type() EventType {
	return s.eventType
}

func NewRemovalEvent(asset Asset) SimpleAssetEvent {
	return SimpleAssetEvent{asset: asset, eventType: Remove}
}

func NewUploadEvent(asset Asset) SimpleAssetEvent {
	return SimpleAssetEvent{asset: asset, eventType: Update}
}
