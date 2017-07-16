package cmd

import (
	"fmt"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Shopify/themekit/kit"
)

func TestVersion(t *testing.T) {
	defer resetArbiter()
	expected := fmt.Sprintf(
		"ThemeKit %s %s/%s\n",
		kit.ThemeKitVersion.String(),
		runtime.GOOS,
		runtime.GOARCH,
	)
	versionCmd.Run(nil, []string{})
	assert.Equal(t, expected, stdOutOutput.String())
}
