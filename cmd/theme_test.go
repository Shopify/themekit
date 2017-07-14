package cmd

import (
	"testing"
)

func TestThemePreRun(t *testing.T) {
	defer resetArbiter()
	// just making sure that it does not throw
	ThemeCmd.PersistentPreRun(nil, []string{})
}

func TestThemePostRun(t *testing.T) {
	defer resetArbiter()
	// just making sure that it does not throw
	ThemeCmd.PersistentPostRun(nil, []string{})
}
