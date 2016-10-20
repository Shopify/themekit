package kit

import (
	"github.com/Shopify/themekit/theme"
)

// EventType is an enum of event types to compare agains event.Type()
type EventType int

type eventCallback func(*ShopifyResponse, Error)

// AssetEvent is an interface that describes events that are related to assets that
// are processed through the eventlog
type AssetEvent struct {
	Asset    theme.Asset
	Type     EventType
	Callback eventCallback
}

const (
	// Create specifies that an AssetEvent is a create event.
	Create EventType = iota
	// Retrieve specifies that an AssetEvent is an update event.
	Retrieve
	// Update specifies that an AssetEvent is an update event.
	Update
	// Remove specifies that an AssetEvent is an delete event.
	Remove
)

func (e EventType) String() string {
	switch e {
	case Create:
		return "Create"
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

func (e EventType) toMethod() string {
	switch e {
	case Retrieve:
		return "GET"
	case Create:
		return "POST"
	case Update:
		return "PUT"
	case Remove:
		return "DELETE"
	default:
		return "Unknown"
	}
}
