package theme

// Theme represents a shopify theme.
type Theme struct {
	Name        string `json:"name"`
	Source      string `json:"src,omitempty"`
	Role        string `json:"role,omitempty"`
	ID          int64  `json:"id,omitempty"`
	Previewable bool   `json:"previewable,omitempty"`
}
