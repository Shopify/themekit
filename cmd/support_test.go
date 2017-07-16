package cmd

import (
	"bytes"
	"log"

	"github.com/Shopify/themekit/kit"
)

var (
	stdOutOutput *bytes.Buffer
	stdErrOutput *bytes.Buffer
)

func init() {
	resetArbiter()
}

func resetArbiter() {
	arbiter = newCommandArbiter()
	arbiter.verbose = true
	arbiter.setFlagConfig()

	stdOutOutput = new(bytes.Buffer)
	stdErrOutput = new(bytes.Buffer)
	stdOut = log.New(stdOutOutput, "", 0)
	stdErr = log.New(stdOutOutput, "", 0)
}

func getClient() (kit.ThemeClient, error) {
	if err := arbiter.generateThemeClients(nil, []string{}); err != nil {
		return kit.ThemeClient{}, err
	}
	return arbiter.activeThemeClients[0], nil
}
