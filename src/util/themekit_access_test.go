package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsThemeKitAccessPassword(t *testing.T) {
	themeKitAccessPassword := "shptka_00000000000000000000000000000000"
	assert.True(t, IsThemeKitAccessPassword(themeKitAccessPassword))

	privateAppPassword := "shp_00000000000000000000000000000000"
	assert.False(t, IsThemeKitAccessPassword(privateAppPassword))
}
