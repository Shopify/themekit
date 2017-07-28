package kit

import (
	"encoding/json"
	"fmt"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRelease_IsValid(t *testing.T) {
	r := Release{}
	assert.False(t, r.IsValid())

	r.Platforms = []Platform{{Name: "test"}}
	assert.True(t, r.IsValid())
}

func TestRelease_IsApplicable(t *testing.T) {
	r := Release{Version: "20.0.0"}
	assert.True(t, r.IsApplicable())

	r = Release{Version: "0.0.0"}
	assert.False(t, r.IsApplicable())

	r = Release{Version: ThemeKitVersion.String()}
	assert.False(t, r.IsApplicable())

	segs := ThemeKitVersion.Segments()
	ver := fmt.Sprintf("v%v.%v.%v", segs[0], segs[1], segs[2])
	r = Release{Version: ver}
	assert.False(t, r.IsApplicable())

	segs = ThemeKitVersion.Segments()
	ver = fmt.Sprintf("v%v.%v.%v", segs[0], segs[1], segs[2]+1)

	r = Release{Version: ver}
	assert.True(t, r.IsApplicable())

	r = Release{Version: ver + "-beta"}
	assert.False(t, r.IsApplicable())

	r = Release{Version: ver + "+prerelease"}
	assert.False(t, r.IsApplicable())

	r = Release{Version: "this_is_not_a_version"}
	assert.False(t, r.IsApplicable())
}

func TestRelease_GetVersion(t *testing.T) {
	r := Release{Version: "20.0.0"}
	assert.Equal(t, "20.0.0", r.GetVersion().String())
}

func TestRelease_ForCurrentPlatform(t *testing.T) {
	thisSystem := runtime.GOOS + "-" + runtime.GOARCH

	r := Release{Version: "20.0.0", Platforms: []Platform{
		{Name: thisSystem},
		{Name: "other-system"},
		{Name: "bad-system"},
	}}
	assert.Equal(t, thisSystem, r.ForCurrentPlatform().Name)

	r = Release{Version: "20.0.0", Platforms: []Platform{
		{Name: "other-system"},
		{Name: "bad-system"},
	}}
	assert.Equal(t, "", r.ForCurrentPlatform().Name)
}

func TestReleasesList_Get(t *testing.T) {
	var releases ReleasesList
	json.Unmarshal([]byte(`[{"version":"0.4.4"},{"version":"0.4.7"},{"version":"0.4.8-prerelease"}]`), &releases)

	r := releases.Get("latest")
	assert.Equal(t, "0.4.7", r.Version)

	r = releases.Get("0.4.4")
	assert.Equal(t, "0.4.4", r.Version)

	r = releases.Get("0.0.0")
	assert.Equal(t, "", r.Version)
	assert.Equal(t, false, r.IsValid())
}

func TestReleasesList_Del(t *testing.T) {
	var releases ReleasesList
	json.Unmarshal([]byte(`[{"version":"0.4.4"},{"version":"0.4.7"}]`), &releases)

	assert.Equal(t, 2, len(releases))
	releases = releases.Del("0.4.7")
	r := releases.Get("0.4.7")
	assert.Equal(t, "", r.Version)
	r = releases.Get("latest")
	assert.Equal(t, "0.4.4", r.Version)
	assert.Equal(t, 1, len(releases))
	releases = releases.Del("0.4.1")
	assert.Equal(t, 1, len(releases))
}
