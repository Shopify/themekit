package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsThemeAccessPassword(t *testing.T) {
	themeKitAccessPassword := "shptka_00000000000000000000000000000000"
	assert.True(t, IsThemeAccessPassword(themeKitAccessPassword))

	privateAppPassword := "shp_00000000000000000000000000000000"
	assert.False(t, IsThemeAccessPassword(privateAppPassword))
}
