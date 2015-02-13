package commands

// import (
// 	"github.com/csaunders/phoenix"
// )

import (
	"fmt"
)

func WatchCommand(args map[string]interface{}) chan bool {
	return Watch()
}

func Watch() (result chan bool) {
	result = make(chan bool)
	fmt.Println("Not yet implemented. Sorry :(")
	close(result)
	return
}
