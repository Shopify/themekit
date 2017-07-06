package cmd

import (
	"testing"
)

func TestThemePreRun(t *testing.T) {
	// just making sure that it does not throw
	ThemeCmd.PersistentPreRun(nil, []string{})
}
