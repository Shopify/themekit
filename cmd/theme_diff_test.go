package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestThemeDiff(t *testing.T) {
	diff := newDiff()
	assert.NotNil(t, newDiff())

	assert.False(t, diff.Any(false))
	diff.Updated = append(diff.Updated, "testupdated.txt")
	assert.True(t, diff.Any(false))
	assert.True(t, diff.Any(true))
	diff.Created = append(diff.Created, "testcreated.txt")
	diff.Removed = append(diff.Removed, "testremoved.txt")

	expected := `Remote files are inconsistent with manifest
Diff:
	New Files:
		- testcreated.txt
	Updated Files:
		- testupdated.txt
	Removed Files:
		- testremoved.txt

You can solve this by running theme download and merging the remote changes
using your favourite diff tool or if you are certain about what you are doing
then use the --force flag
`
	assert.Equal(t, expected, diff.Error())
}
