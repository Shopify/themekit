package kit

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEventType(t *testing.T) {
	assert.Equal(t, "Create", Create.String())
	assert.Equal(t, "Retrieve", Retrieve.String())
	assert.Equal(t, "Update", Update.String())
	assert.Equal(t, "Remove", Remove.String())
	assert.Equal(t, "Unknown", EventType(99).String())

	assert.Equal(t, "POST", Create.toMethod())
	assert.Equal(t, "GET", Retrieve.toMethod())
	assert.Equal(t, "PUT", Update.toMethod())
	assert.Equal(t, "DELETE", Remove.toMethod())
	assert.Equal(t, "Unknown", EventType(99).toMethod())
}
