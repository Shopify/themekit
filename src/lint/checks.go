package lint

import (
	"fmt"
)

type EventType int

const (
	Start EventType = iota
	Filter
	Tag
	Node
	Assign
	BeginDoc
	EndDoc
	BeginBlock
	EndBlock
	End
)

type Check func(evtType EventType)
type CheckCollection []Check

var AllChecks = CheckCollection([]Check{PrintCheck})

var PrintCheck = func(evtType EventType) {
	fmt.Println(evtType)
}

func (col CheckCollection) Call(evtType EventType) {
	for _, check := range col {
		check(evtType)
	}
}
