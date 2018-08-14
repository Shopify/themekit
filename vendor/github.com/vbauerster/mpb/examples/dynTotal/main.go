package main

import (
	"io"
	"math/rand"
	"time"

	"github.com/vbauerster/mpb"
	"github.com/vbauerster/mpb/decor"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func main() {
	p := mpb.New(mpb.WithWidth(64))

	var total int64
	bar := p.AddBar(total,
		mpb.PrependDecorators(decor.Counters(decor.UnitKiB, "% .1f / % .1f")),
		mpb.AppendDecorators(decor.Percentage()),
	)

	maxSleep := 100 * time.Millisecond
	read := makeStream(200)
	for {
		n, err := read()
		total += int64(n)
		time.Sleep(time.Duration(rand.Intn(10)+1) * maxSleep / 10)
		bar.IncrBy(n)
		if err == io.EOF {
			// total is known, final=true
			bar.SetTotal(total, true)
			break
		}
		// total is unknown, final=false
		bar.SetTotal(total+2048, false)
	}

	p.Wait()
}

func makeStream(limit int) func() (int, error) {
	return func() (int, error) {
		if limit <= 0 {
			return 0, io.EOF
		}
		limit--
		return rand.Intn(1024) + 1, nil
	}
}
