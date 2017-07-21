package kit

import (
	"testing"

	"github.com/fsnotify/fsnotify"
	"github.com/stretchr/testify/assert"
)

func TestGet(t *testing.T) {
	eventMap := newEventMap()
	eventMap.New("test")
	_, ok := eventMap.Get("test")
	assert.True(t, ok)
	_, ok = eventMap.Get("nope")
	assert.False(t, ok)
}

func TestDel(t *testing.T) {
	eventMap := newEventMap()
	eventMap.New("test")
	_, ok := eventMap.Get("test")
	assert.True(t, ok)
	eventMap.Del("test")
	_, ok = eventMap.Get("test")
	assert.False(t, ok)
}

func TestCount(t *testing.T) {
	eventMap := newEventMap()
	eventMap.New("test")
	eventMap.New("test2")
	eventMap.New("test")
	eventMap.New("test3")
	assert.Equal(t, eventMap.Count(), 3)
}

func TestSet(t *testing.T) {
	eventMap := newEventMap()
	myChan := make(chan fsnotify.Event)
	eventMap.Set("test", myChan)
	assert.Equal(t, eventMap.Count(), 1)
	getChan, _ := eventMap.Get("test")
	assert.Equal(t, getChan, myChan)
}

func TestNew(t *testing.T) {
	eventMap := newEventMap()
	myChan := eventMap.New("test")
	assert.Equal(t, eventMap.Count(), 1)
	getChan, _ := eventMap.Get("test")
	assert.Equal(t, getChan, myChan)
}
