package kit

// Theme represents a shopify theme.
type Theme struct {
	ID          int64  `json:"id,omitempty"`
	Name        string `json:"name"`
	Source      string `json:"src,omitempty"`
	Role        string `json:"role,omitempty"`
	Previewable bool   `json:"previewable,omitempty"`
	Processing  bool   `json:"processing,omitempty"`
}
