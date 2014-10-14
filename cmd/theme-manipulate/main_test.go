package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestPermittedCommands(t *testing.T) {
	expected :=
		`An operation to be performed against the theme.
  Valid commands are:
    upload: Add file(s) to theme [default]
    download: Download file(s) from theme
    remove: Remove file(s) from theme
    replace: Overwrite theme file(s)`

	actual := CommandDescription("upload")
	assert.Equal(t, expected, actual)
}
