package commands

import (
	"fmt"
	"github.com/csaunders/phoenix"
	"log"
	"net/http"
	"os"
)

const (
	MasterBranch   = "master"
	LatestRelease  = "latest"
	ThemeZipRoot   = "https://github.com/Shopify/Timber/archive/"
	TimberFeedPath = "https://github.com/Shopify/Timber/releases.atom"
)

func BootstrapCommand(args map[string]interface{}) (done chan bool) {
	return
}

func Bootstrap(client phoenix.ThemeClient, version string) (done chan bool) {
	var zipLocation string
	if version == MasterBranch {
		zipLocation = zipPath(MasterBranch)
	} else {
		zipLocation = zipPathForVersion(version)
	}
	clientForNewTheme, wg := client.CreateTheme(version, zipLocation)
	wg.Wait()

	return Download(clientForNewTheme, []string{})
}

func zipPath(version string) string {
	return ThemeZipRoot + version + ".zip"
}

func zipPathForVersion(version string) string {
	feed := downloadAtomFeed()
	entry := findReleaseWith(feed, version)
	return zipPath(entry.Title)
}

func downloadAtomFeed() phoenix.Feed {
	resp, err := http.Get(TimberFeedPath)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	feed, err := phoenix.LoadFeed(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	return feed
}

func findReleaseWith(feed phoenix.Feed, version string) phoenix.Entry {
	if version == LatestRelease {
		return feed.LatestEntry()
	}
	for _, entry := range feed.Entries {
		if entry.Title == version {
			return entry
		}
	}
	logAndDie(feed, version)
	return phoenix.Entry{}
}

func logAndDie(feed phoenix.Feed, version string) {
	fmt.Println(phoenix.RedText("Invalid Timber Version - " + version))
	fmt.Println("Available Versions Are:")
	fmt.Println("  - master")
	fmt.Println("  - latest")
	for _, entry := range feed.Entries {
		fmt.Println("  - " + entry.Title)
	}
	os.Exit(1)

}
