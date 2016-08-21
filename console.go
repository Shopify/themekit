package themekit

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"os"
	"time"
)

// Console ... TODO rename
type Console struct {
	writer     *bufio.Writer
	done       chan bool
	timeout    time.Duration
	lastTickAt time.Time
}

// Initialize ... TODO
func (c *Console) Initialize() {
	c.writer = bufio.NewWriter(os.Stdout)
	c.lastTickAt = time.Now()
}

// Write ... TODO
func (c *Console) Write(str string) {
	c.writeAndFlush(fmt.Sprintf("%s\n", str))
	c.lastTickAt = time.Now()
}

// WriteFatalError ... TODO
func (c *Console) WriteFatalError(e error) {
	libraryInfo := fmt.Sprintf("%s%s%s", MessageSeparator, LibraryInfo(), MessageSeparator)
	c.writeAndFlush(RedText(fmt.Sprintf("%s\n", errors.New(libraryInfo))))
	log.Fatal(e)
}

// HandleTimeout ... TODO
func (c *Console) HandleTimeout(timeout time.Duration, done chan bool) {
	c.timeout = timeout
	c.done = done

	go func() {
		timedout := false
		for {
			select {
			case <-time.Tick(c.timeout):
				if time.Now().After(c.lastTickAt.Add(c.timeout)) {
					close(c.done)
					timedout = true
				}
			}

			if timedout {
				break
			}
		}
	}()
}

func (c *Console) writeAndFlush(str string) {
	c.writer.WriteString(str)
	c.writer.Flush()
}
