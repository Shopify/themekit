package kit

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type VersionTestSuite struct {
	suite.Suite
	version version
}

func (suite *VersionTestSuite) SetupTest() {
	suite.version = version{1, 52, 99}
}

func (suite *VersionTestSuite) TestComparingDifferentVersions() {
	tests := []struct {
		version  version
		expected versionComparisonResult
	}{
		{version{1, 51, 0}, VersionLessThan},
		{version{0, 0, 0}, VersionLessThan},
		{version{2, 0, 0}, VersionGreaterThan},
		{version{1, 53, 99}, VersionGreaterThan},
		{version{1, 52, 100}, VersionGreaterThan},
		{version{1, 52, 99}, VersionEqual},
	}
	for _, test := range tests {
		assert.Equal(suite.T(), test.expected, test.version.Compare(suite.version.String()))
	}
}

func (suite *VersionTestSuite) TestStringifyingAVersion() {
	assert.Equal(suite.T(), "v1.52.99", suite.version.String())
}

func (suite *VersionTestSuite) TestParsingAVersionString() {
	assert.Equal(suite.T(), VersionEqual, suite.version.Compare("1.52.99"))
}

func (suite *VersionTestSuite) TestParsingAVersionStringWithPrefixedV() {
	assert.Equal(suite.T(), VersionEqual, suite.version.Compare("v1.52.99"))
}

func TestVersionTestSuite(t *testing.T) {
	suite.Run(t, new(VersionTestSuite))
}
