package kit

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLibraryInfo(t *testing.T) {
	messageSeparator := "\n----------------------------------------------------------------\n"
	info := fmt.Sprintf("\t%s %s", "ThemeKit - Shopify Theme Utilities", ThemeKitVersion.String())
	assert.Equal(t, fmt.Sprintf("%s%s%s", messageSeparator, info, messageSeparator), LibraryInfo())
}
