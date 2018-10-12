package main

import "github.com/Shopify/themekit/src/static"

func main() {
	if err := static.Bundle("theme-template", "cmd/static/generated-assets.go"); err != nil {
		panic(err)
	}
}
