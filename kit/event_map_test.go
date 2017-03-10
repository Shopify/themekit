package kit

import (
	"testing"

	"github.com/fsnotify/fsnotify"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type EventMapTestSuite struct {
	suite.Suite
}

func (suite *EventMapTestSuite) TestGet() {
	eventMap := newEventMap()
	eventMap.New("test")
	_, ok := eventMap.Get("test")
	assert.True(suite.T(), ok)
	_, ok = eventMap.Get("nope")
	assert.False(suite.T(), ok)
}

func (suite *EventMapTestSuite) TestDel() {
	eventMap := newEventMap()
	eventMap.New("test")
	_, ok := eventMap.Get("test")
	assert.True(suite.T(), ok)
	eventMap.Del("test")
	_, ok = eventMap.Get("test")
	assert.False(suite.T(), ok)
}

func (suite *EventMapTestSuite) TestCount() {
	eventMap := newEventMap()
	eventMap.New("test")
	eventMap.New("test2")
	eventMap.New("test")
	eventMap.New("test3")
	assert.Equal(suite.T(), eventMap.Count(), 3)
}

func (suite *EventMapTestSuite) TestSet() {
	eventMap := newEventMap()
	myChan := make(chan fsnotify.Event)
	eventMap.Set("test", myChan)
	assert.Equal(suite.T(), eventMap.Count(), 1)
	getChan, _ := eventMap.Get("test")
	assert.Equal(suite.T(), getChan, myChan)
}

func (suite *EventMapTestSuite) TestNew() {
	eventMap := newEventMap()
	myChan := eventMap.New("test")
	assert.Equal(suite.T(), eventMap.Count(), 1)
	getChan, _ := eventMap.Get("test")
	assert.Equal(suite.T(), getChan, myChan)
}

func TestEventMapTestSuite(t *testing.T) {
	suite.Run(t, new(EventMapTestSuite))
}
