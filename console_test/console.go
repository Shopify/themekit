package main

import (
	"fmt"
	"time"

	"github.com/Shopify/themekit"
)

func main() {
	done := make(chan bool)
	// console := themekit.NewConsole()
	console := new(themekit.Console)
	console.Initialize()

	fmt.Printf("STARTING: %v", time.Now())

	console.HandleTimeout(5*time.Second, done)

	go func() {
		console.Write("Immediately")

		time.Sleep(1 * time.Second)
		console.Write("After 1 Second")

		time.Sleep(2 * time.Second)
		console.Write("After 2 Second")

		time.Sleep(3 * time.Second)
		console.Write("After 3 Second")

		time.Sleep(4 * time.Second)
		console.Write("After 4 Second")

		time.Sleep(8 * time.Second)
		console.Write("After 8 Second")
	}()

	<-done

	fmt.Println("Done!")
}
