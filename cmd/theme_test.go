package cmd

import (
	"bytes"
	"log"
	"testing"

	"github.com/Shopify/themekit/kit"
	"github.com/Shopify/themekit/kittest"
)

var (
	stdOutOutput *bytes.Buffer
	stdErrOutput *bytes.Buffer
	logLock      sync.Mutex
)

func resetLog() {
	stdOutOutput = new(bytes.Buffer)
	stdErrOutput = new(bytes.Buffer)
	stdOut = log.New(stdOutOutput, "", 0)
	stdErr = log.New(stdOutOutput, "", 0)
}

func TestThemePreRun(t *testing.T) {
	server := kittest.NewTestServer()
	defer server.Close()
	defer resetArbiter()

	kit.ThemeKitReleasesURL = server.URL + "/themekit_update"

	// just making sure that it does not throw
	ThemeCmd.PersistentPreRun(nil, []string{})
}

func TestThemePostRun(t *testing.T) {
	defer resetArbiter()
	// just making sure that it does not throw
	ThemeCmd.PersistentPostRun(nil, []string{})
}
