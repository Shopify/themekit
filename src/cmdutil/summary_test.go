package cmdutil

import (
	"bytes"
	"fmt"
	"log"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Shopify/themekit/src/env"
	"github.com/Shopify/themekit/src/file"
)

func TestSummaryCompleteOp(t *testing.T) {
	summary := cmdSummary{}

	summary.completeOp(file.Get)
	assert.Equal(t, summary.downloaded, int32(1))

	summary.completeOp(file.Update)
	assert.Equal(t, summary.uploaded, int32(1))

	summary.completeOp(file.Skip)
	assert.Equal(t, summary.skipped, int32(1))

	summary.completeOp(file.Remove)
	assert.Equal(t, summary.removed, int32(1))

	assert.Equal(t, summary.actions, int32(4))
}

func TestSummaryDisable(t *testing.T) {
	summary := cmdSummary{}
	assert.False(t, summary.disabled)
	summary.disable()
	assert.True(t, summary.disabled)
}

func TestSummaryErr(t *testing.T) {
	summary := cmdSummary{}
	assert.Equal(t, summary.errors, []string(nil))
	summary.err("no good")
	assert.Equal(t, summary.errors, []string{"no good"})
}

func TestSummaryHasErrors(t *testing.T) {
	summary := cmdSummary{}
	assert.False(t, summary.hasErrors())
	summary.err("no good")
	assert.True(t, summary.hasErrors())
	summary.disable()
	assert.False(t, summary.hasErrors())
}

func TestSummaryDisplay(t *testing.T) {
	out, err := rundisplay(cmdSummary{actions: 23})
	assert.Equal(t, out, fmt.Sprintf("[sum] 23 files\n"))
	assert.Equal(t, err, "")

	out, err = rundisplay(cmdSummary{actions: 23, downloaded: 21})
	assert.Equal(t, out, fmt.Sprintf("[sum] 23 files, Downloaded: 21\n"))
	assert.Equal(t, err, "")

	out, err = rundisplay(cmdSummary{actions: 23, uploaded: 21})
	assert.Equal(t, out, fmt.Sprintf("[sum] 23 files, Updated: 21\n"))
	assert.Equal(t, err, "")

	out, err = rundisplay(cmdSummary{actions: 23, removed: 42})
	assert.Equal(t, out, fmt.Sprintf("[sum] 23 files, Removed: 42\n"))
	assert.Equal(t, err, "")

	out, err = rundisplay(cmdSummary{actions: 23, skipped: 11})
	assert.Equal(t, out, fmt.Sprintf("[sum] 23 files, No Change: 11\n"))
	assert.Equal(t, err, "")

	out, err = rundisplay(cmdSummary{actions: 23, errors: []string{"one", "two", "three"}})
	assert.Equal(t, out, fmt.Sprintf("[sum] 23 files, Errored: 3\n"))
	assert.Equal(t, err, "[sum] Errors encountered: \n\tone\n\ttwo\n\tthree\n")
}

func rundisplay(summary cmdSummary) (stdout, stderr string) {
	stdOut := bytes.NewBufferString("")
	stdErr := bytes.NewBufferString("")
	ctx := &Ctx{Env: &env.Env{Name: "sum"}, Log: log.New(stdOut, "", 0), ErrLog: log.New(stdErr, "", 0)}
	summary.display(ctx)
	return stdOut.String(), stdErr.String()
}
