package main

import (
	"flag"
	"os"

	"github.com/Shopify/themekit/src/colors"
	"github.com/Shopify/themekit/src/release"
)

var (
	stdOut = colors.ColorStdOut
	stdErr = colors.ColorStdErr
)

func main() {
	var force, destroy bool
	var key, secret string
	flag.BoolVar(&force, "f", false, "Skip checks of versions. Useful for updating a deploy")
	flag.BoolVar(&destroy, "d", false, "Destroy release version. This will remove a version from the release feed.")
	flag.StringVar(&key, "k", "", "Amazon s3 key")
	flag.StringVar(&secret, "s", "", "Amazon s3 secret")
	flag.Parse()

	if len(flag.Args()) <= 0 {
		stdErr.Println(colors.Red("please provide a version number"))
		os.Exit(0)
	}

	var err error
	if destroy {
		err = release.Remove(key, secret, flag.Args()[0])
	} else {
		err = release.Update(key, secret, flag.Args()[0], force)
	}

	if err != nil {
		stdErr.Println(colors.Red(err.Error()))
		os.Exit(0)
	}

	stdOut.Println(colors.Green("Deploy succeeded"))
}
