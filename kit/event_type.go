package kit

import (
	"github.com/Shopify/themekit/theme"
)

// EventType is an enum of event types to compare agains event.Type()
type EventType int

// AssetEvent is an interface that describes events that are related to assets that
// are processed through the eventlog
type AssetEvent interface {
	Asset() theme.Asset
	Type() EventType
}

const (
	// Retrieve specifies that an AssetEvent is an update event.
	Retrieve EventType = iota
	// Update specifies that an AssetEvent is an update event.
	Update
	// Remove specifies that an AssetEvent is an delete event.
	Remove
)

func (e EventType) String() string {
	switch e {
	case Retrieve:
		return "Retrieve"
	case Update:
		return "Update"
	case Remove:
		return "Remove"
	default:
		return "Unknown"
	}
}

func (e EventType) ToMethod() string {
	switch e {
	case Retrieve:
		return "GET"
	case Update:
		return "POST"
	case Remove:
		return "DELETE"
	default:
		return "Unknown"
	}
}
