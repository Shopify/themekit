package kit

import (
	"sync"

	"github.com/fsnotify/fsnotify"
)

type eventMap struct {
	sync.RWMutex
	events map[string]chan fsnotify.Event
}

func newEventMap() *eventMap {
	return &eventMap{
		events: map[string]chan fsnotify.Event{},
	}
}

func (em *eventMap) Get(eventName string) (chan fsnotify.Event, bool) {
	em.RLock()
	defer em.RUnlock()
	eventsChan, ok := em.events[eventName]
	return eventsChan, ok
}

func (em *eventMap) Del(eventName string) {
	em.Lock()
	defer em.Unlock()
	delete(em.events, eventName)
}

func (em *eventMap) Count() int {
	em.RLock()
	defer em.RUnlock()
	return len(em.events)
}

func (em *eventMap) Set(eventName string, eventsChan chan fsnotify.Event) {
	em.Lock()
	defer em.Unlock()
	em.events[eventName] = eventsChan
}

func (em *eventMap) New(eventName string) chan fsnotify.Event {
	eventsChan := make(chan fsnotify.Event)
	em.Set(eventName, eventsChan)
	return eventsChan
}
