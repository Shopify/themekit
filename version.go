package themekit

import (
	"crypto"
	"encoding/hex"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/inconshreveable/go-update"
)

// TKVersion ... TODO
var TKVersion = Version{Major: 0, Minor: 4, Patch: 0}

// ThemeKitVersion ... TODO
var ThemeKitVersion = TKVersion.String()

// VersionComparisonResult ... TODO
type VersionComparisonResult int

const (
	// VersionLessThan .. TODO
	VersionLessThan VersionComparisonResult = -1
	// VersionEqual ... TODO
	VersionEqual = 0
	// VersionGreaterThan ... TODO
	VersionGreaterThan = 1
)

// LibraryInfo ... TODO
func LibraryInfo() []string {
	return []string{
		"ThemeKit - Shopify Theme Utilities",
		ThemeKitVersion,
		"Author: Chris Saunders",
	}
}

// Version ... TODO
type Version struct {
	Major int
	Minor int
	Patch int
}

func (v Version) String() string {
	return fmt.Sprintf("v%d.%d.%d", v.Major, v.Minor, v.Patch)
}

func (v Version) toArray() [3]int {
	return [3]int{v.Major, v.Minor, v.Patch}
}

// Compare ... I often get confused by comparison, so comparison results are going
// to be the same as what <=> would return in Ruby.
// http://ruby-doc.org/core-1.9.3/Comparable.html
func (v Version) Compare(o Version) VersionComparisonResult {
	vAry := v.toArray()
	oAry := o.toArray()
	for i := 0; i < len(vAry); i++ {
		diff := vAry[i] - oAry[i]
		if diff < 0 {
			return VersionLessThan
		} else if diff > 0 {
			return VersionGreaterThan
		}
	}
	return VersionEqual
}

// ParseVersionString ... TODO
func ParseVersionString(ver string) Version {
	sanitizedVer := strings.Replace(ver, "v", "", 1)
	expandedVersionString := strings.Split(sanitizedVer, ".")
	major, _ := strconv.Atoi(expandedVersionString[0])
	minor, _ := strconv.Atoi(expandedVersionString[1])
	patch, _ := strconv.Atoi(expandedVersionString[2])
	return Version{Major: major, Minor: minor, Patch: patch}
}

// ApplyUpdate ... TODO
func ApplyUpdate(updateURL, digest string) error {
	checksum, err := hex.DecodeString(digest)
	if err != nil {
		return err
	}

	updateFile, err := http.Get(updateURL)
	if err != nil {
		return err
	}
	defer updateFile.Body.Close()

	err = update.Apply(updateFile.Body, update.Options{
		Hash:     crypto.MD5,
		Checksum: checksum,
	})
	if err != nil {
		if rerr := update.RollbackError(err); rerr != nil {
			fmt.Printf("Failed to rollback from bad update: %v", rerr)
		}
	}
	return err
}
