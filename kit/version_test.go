package kit

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestComparingDifferentVersions(t *testing.T) {
	tests := []struct {
		me       version
		other    version
		expected versionComparisonResult
	}{
		{version{1, 0, 0}, version{1, 0, 1}, VersionLessThan},
		{version{1, 0, 0}, version{0, 9, 9}, VersionGreaterThan},
		{version{1, 0, 0}, version{1, 0, 0}, VersionEqual},
	}
	for _, test := range tests {
		assert.Equal(t, test.expected, test.me.Compare(test.other))
	}
}

func TestStringifyingAVersion(t *testing.T) {
	assert.Equal(t, "v1.0.0", version{1, 0, 0}.String())
}

func TestParsingAVersionString(t *testing.T) {
	expected := version{1, 52, 99}
	actual := ParseVersionString("1.52.99")
	assert.Equal(t, VersionEqual, expected.Compare(actual))
}

func TestParsingAVersionStringWithPrefixedV(t *testing.T) {
	expected := version{1, 52, 99}
	actual := ParseVersionString("v1.52.99")
	assert.Equal(t, VersionEqual, expected.Compare(actual))
}
