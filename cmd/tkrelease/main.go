package main

import (
	"flag"
	"os"

	"github.com/Shopify/themekit/src/colors"
	"github.com/Shopify/themekit/src/release"
)

var (
	force   bool
	destroy bool
	stdOut  = colors.ColorStdOut
	stdErr  = colors.ColorStdErr
)

func init() {
	flag.BoolVar(&force, "f", false, "Skip checks of versions. Useful for updating a deploy")
	flag.BoolVar(&destroy, "d", false, "Destroy release version. This will remove a version from the release feed.")
	flag.Parse()
}

func main() {
	if len(flag.Args()) > 0 {
		stdErr.Println(colors.Red("please provide a version number"))
		os.Exit(0)
	}

	var err error
	if destroy {
		err = release.Remove(flag.Args()[0])
	} else {
		err = release.Update(flag.Args()[0], force)
	}

	if err != nil {
		stdErr.Println(colors.Red(err.Error()))
		os.Exit(0)
	}

	stdOut.Println(colors.Green("Deploy succeeded"))
}
