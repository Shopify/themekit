package main

import (
	"bytes"
	"crypto/md5"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/fatih/color"
	"github.com/hashicorp/go-version"
	"github.com/mattn/go-colorable"

	"github.com/Shopify/themekit/kit"
)

const (
	allReleasesFilename   = "releases/all.json"
	latestReleaseFilename = "releases/latest.json"
	region                = "us-east-1"
	bucketName            = "shopify-themekit"
)

var (
	force   bool
	destroy bool
	distDir = filepath.Join("build", "dist")
	builds  = map[string]string{
		"darwin-amd64":  "theme",
		"linux-386":     "theme",
		"linux-amd64":   "theme",
		"windows-386":   "theme.exe",
		"windows-amd64": "theme.exe",
	}
	uploader *s3manager.Uploader
	red      = color.New(color.FgRed).SprintFunc()
	green    = color.New(color.FgGreen).SprintFunc()
	stdOut   = log.New(colorable.NewColorableStdout(), "", 0)
	stdErr   = log.New(colorable.NewColorableStderr(), "", 0)
)

type deploySecrets struct {
	Key    string `json:"key"`
	Secret string `json:"secret"`
}

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
	assert(len(flag.Args()) > 0, "not enough args")
	ver := flag.Args()[0]

	if !force {
		requestedVersion, err := version.NewVersion(ver)
		checkErr(err)
		assert(requestedVersion.Equal(kit.ThemeKitVersion), "Deploy version does not match themekit verison")
	}

	checkErr(buildS3Uploader())

	_, err := os.Stat(distDir)
	assert(!os.IsNotExist(err), "Dist folder does not exist. Run 'make dist' before attempting to create a new release")
	checkErr(err)

	if destroy {
		checkErr(removeRelease(ver))
	} else {
		release, err := updateRelease(ver)
		checkErr(err)
		checkErr(updateLatest(release))
	}

	stdOut.Println(green("Deploy succeeded"))
}

func buildS3Uploader() error {
	stdOut.Println(green("Connecting to S3"))
	raw, err := ioutil.ReadFile(".env")
	if err != nil {
		return err
	}

	var secrets deploySecrets
	err = json.Unmarshal(raw, &secrets)
	if err != nil {
		return err
	}

	creds := credentials.NewStaticCredentials(secrets.Key, secrets.Secret, "")
	cfg := aws.NewConfig().WithRegion(region).WithCredentials(creds)
	uploader = s3manager.NewUploader(session.New(cfg))

	return nil
}

func buildRelease(version string) (kit.Release, error) {
	platforms := []kit.Platform{}

	stdOut.Println("building release for", green(version))
	for platformName, binName := range builds {
		platform, _ := buildPlatform(version, platformName, binName)
		platforms = append(platforms, platform)
	}

	return kit.Release{Version: version, Platforms: platforms}, nil
}

func buildPlatform(version, platformName, binName string) (kit.Platform, error) {
	f, err := os.Open(filepath.Join(distDir, platformName, binName))
	if err != nil {
		return kit.Platform{}, err
	}
	defer f.Close()

	data, err := ioutil.ReadAll(f)
	if err != nil {
		return kit.Platform{}, err
	}

	fullName := strings.Join([]string{version, platformName, binName}, "/")
	url, err := uploadFile(fullName, f)
	if err != nil {
		return kit.Platform{}, err
	}

	stdOut.Println("uploading", green(platformName))
	return kit.Platform{
		Name:   platformName,
		URL:    url,
		Digest: fmt.Sprintf("%x", md5.Sum(data)),
	}, nil
}

func uploadFile(fileName string, body io.ReadSeeker) (string, error) {
	resp, err := uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(bucketName),
		ACL:    aws.String("public-read"),
		Key:    aws.String(fileName),
		Body:   body,
	})
	if err != nil {
		return "", err
	}
	fileURL, err := url.QueryUnescape(resp.Location)
	if err != nil {
		return "", err
	}
	return fileURL, nil
}

func uploadJSON(filename string, data interface{}) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}
	_, err = uploadFile(filename, bytes.NewReader(jsonData))
	return err
}

func removeRelease(ver string) error {
	stdOut.Printf("Removing release %s", green(ver))
	releases, err := kit.FetchReleases()
	if err != nil {
		return err
	}

	requestedRelease := releases.Get(ver)
	assert(requestedRelease.IsValid(), "version has not be deployed.")

	releases = releases.Del(ver)
	err = uploadJSON(allReleasesFilename, releases)
	if err != nil {
		return err
	}
	stdOut.Println("Updated release feed")

	currentLatest, err := kit.FetchLatest()
	if err != nil {
		return err
	}
	if currentLatest.GetVersion().Equal(requestedRelease.GetVersion()) {
		latestVersion := releases.Get("latest")
		stdOut.Println("Updating latest to", green(latestVersion.Version))
		return uploadJSON(latestReleaseFilename, latestVersion)
	}
	return nil
}

func updateRelease(ver string) (kit.Release, error) {
	releases, err := kit.FetchReleases()
	if err != nil {
		return kit.Release{}, err
	}

	if !force {
		requestedRelease := releases.Get(ver)
		assert(!requestedRelease.IsValid(), "version has already been deployed.")
	}

	release, err := buildRelease(ver)
	if err != nil {
		return kit.Release{}, err
	}

	defer stdOut.Println("Updated release feed")
	releases = append(releases.Del(ver), release)

	return release, uploadJSON(allReleasesFilename, releases)
}

func updateLatest(release kit.Release) error {
	currentLatest, err := kit.FetchLatest()
	if err != nil {
		return err
	}
	releaseVersion := release.GetVersion()
	if currentLatest.GetVersion().LessThan(releaseVersion) && releaseVersion.Metadata() == "" && releaseVersion.Prerelease() == "" {
		stdOut.Println("Updating latest to", green(release.Version))
		return uploadJSON(latestReleaseFilename, release)
	}
	return nil
}
