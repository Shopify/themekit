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

	expected := `Unexpected changes made on remote.
Diff:
	New Files:
		- testcreated.txt
	Updated Files:
		- testupdated.txt
	Removed Files:
		- testremoved.txt

You can solve this by running 'theme download' to get the most recent copy of these files.
Running 'theme download' will overwrite any changes you have made so make sure your work is
commited to your VCS before doing so.

If you are certain that you want to overwrite any changes then use the --force flag
`
	assert.Equal(t, expected, diff.Error())
}
