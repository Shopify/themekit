package cmd

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/Shopify/themekit/cmd/internal/atom"
)

type BootstrapTestSuite struct {
	suite.Suite
}

func (suite *BootstrapTestSuite) TestBootstrap() {
}

func (suite *BootstrapTestSuite) TestZipPath() {
	assert.Equal(suite.T(), themeZipRoot+"foo.zip", zipPath("foo"))
}

func (suite *BootstrapTestSuite) TestZipPathForVersion() {
}

func (suite *BootstrapTestSuite) TestDownloadAtomFeed() {
}

func (suite *BootstrapTestSuite) TestFindReleaseWith() {
	feed := loadAtom()
	entry, err := findReleaseWith(feed, "latest")
	assert.Equal(suite.T(), feed.LatestEntry(), entry)
	assert.Nil(suite.T(), err)

	entry, err = findReleaseWith(feed, "v2.0.2")
	assert.Equal(suite.T(), "v2.0.2", entry.Title)
	assert.Nil(suite.T(), err)

	entry, err = findReleaseWith(feed, "nope")
	assert.Equal(suite.T(), "Invalid Feed", entry.Title)
	assert.NotNil(suite.T(), err)
}

func (suite *BootstrapTestSuite) TestBuildInvalidVersionError() {
	feed := loadAtom()
	err := buildInvalidVersionError(feed, "nope")
	assert.Equal(suite.T(), "Invalid Timber Version: nope\nAvailable Versions Are:\n- master\n- latest\n- v2.0.2\n- v2.0.1\n- v2.0.0\n- v1.3.2\n- v1.3.1\n- v1.3.0\n- v1.2.1\n- v1.2.0\n- v1.1.3\n- v1.1.2\n- v1.1.1\n- v1.1.0\n- v1.0.0\n", err.Error())
}

func TestBootstrapTestSuite(t *testing.T) {
	suite.Run(t, new(BootstrapTestSuite))
}

func loadAtom() atom.Feed {
	stream, _ := os.Open("../fixtures/releases.atom")
	feed, _ := atom.LoadFeed(stream)
	return feed
}
