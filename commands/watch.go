package commands

import (
	"github.com/com/csaunders/phoenix"
)

func WatchCommand(args map[string]interface{}) chan bool {
	return Watch()
}

func Watch() (result chan bool) {
	return
}
