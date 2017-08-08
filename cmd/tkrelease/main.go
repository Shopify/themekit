package main

import (
	"flag"
	"log"
	"os"
	"path/filepath"

	"github.com/fatih/color"
	"github.com/hashicorp/go-version"
	"github.com/mattn/go-colorable"

	"github.com/Shopify/themekit/kit"
)

const (
	allReleasesFilename   = "releases/all.json"
	latestReleaseFilename = "releases/latest.json"
)

var (
	force   bool
	destroy bool
	distDir = filepath.Join("build", "dist")
	red     = color.New(color.FgRed).SprintFunc()
	green   = color.New(color.FgGreen).SprintFunc()
	stdOut  = log.New(colorable.NewColorableStdout(), "", 0)
	stdErr  = log.New(colorable.NewColorableStderr(), "", 0)
)

func init() {
	flag.BoolVar(&force, "f", false, "Skip checks of versions. Useful for updating a deploy")
	flag.BoolVar(&destroy, "d", false, "Destroy release version. This will remove a version from the release feed.")
	flag.Parse()
}

func checkErr(err error) {
	if err != nil {
		stdErr.Println(red(err.Error()))
		os.Exit(0)
	}
}

func assert(must bool, message string) {
	if !must {
		stdErr.Println(red(message))
		os.Exit(0)
	}
}

func main() {
	assert(len(flag.Args()) > 0, "please provide a version number")
	ver := flag.Args()[0]

	if destroy {
		removeRelease(ver)
	} else {
		updateRelease(ver)
	}

	stdOut.Println(green("Deploy succeeded"))
}

func removeRelease(ver string) {
	stdOut.Printf("Removing release %s", green(ver))
	releases, err := kit.FetchReleases()
	checkErr(err)

	requestedRelease := releases.Get(ver)
	assert(requestedRelease.IsValid(), "version has not be deployed.")

	uploader, err := newS3Uploader()
	checkErr(err)

	releases = releases.Del(ver)
	checkErr(uploader.json(allReleasesFilename, releases))

	currentLatest, err := kit.FetchLatest()
	checkErr(err)
	if currentLatest.GetVersion().Equal(requestedRelease.GetVersion()) {
		latestVersion := releases.Get("latest")
		stdOut.Println("Updating latest to", green(latestVersion.Version))
		checkErr(uploader.json(latestReleaseFilename, latestVersion))
	}
}

func updateRelease(ver string) {
	if !force {
		requestedVersion, _ := version.NewVersion(ver)
		assert(requestedVersion.Equal(kit.ThemeKitVersion), "Deploy version does not match themekit version")
	}

	_, err := os.Stat(distDir)
	assert(!os.IsNotExist(err), "Dist folder does not exist. Run 'make dist' before attempting to create a new release")

	releases, err := kit.FetchReleases()
	checkErr(err)

	if !force {
		requestedRelease := releases.Get(ver)
		assert(!requestedRelease.IsValid(), "version has already been deployed.")
	}

	uploader, err := newS3Uploader()
	checkErr(err)

	stdOut.Println("building release for", green(ver))
	release, err := buildRelease(ver, uploader)
	checkErr(err)

	releases = append(releases.Del(ver), release)

	checkErr(uploader.json(allReleasesFilename, releases))

	currentLatest, err := kit.FetchLatest()
	checkErr(err)

	releaseVersion := release.GetVersion()
	if currentLatest.GetVersion().LessThan(releaseVersion) && releaseVersion.Metadata() == "" && releaseVersion.Prerelease() == "" {
		stdOut.Println("Updating latest to", green(release.Version))
		checkErr(uploader.json(latestReleaseFilename, release))
	}
}
