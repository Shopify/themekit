package kit

import (
	"encoding/json"
	"fmt"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReleaseIsValid(t *testing.T) {
	r := release{}
	assert.Equal(t, false, r.IsValid())

	r.Platforms = []platform{{Name: "test"}}
	assert.Equal(t, true, r.IsValid())
}

func TestReleaseIsApplicable(t *testing.T) {
	r := release{Version: "20.0.0"}
	assert.Equal(t, true, r.IsApplicable())

	r = release{Version: "0.0.0"}
	assert.Equal(t, false, r.IsApplicable())

	r = release{Version: ThemeKitVersion.String()}
	assert.Equal(t, false, r.IsApplicable())

	segs := ThemeKitVersion.Segments()
	ver := fmt.Sprintf("v%v.%v.%v", segs[0], segs[1], segs[2])
	r = release{Version: ver}
	assert.Equal(t, false, r.IsApplicable())

	segs = ThemeKitVersion.Segments()
	ver = fmt.Sprintf("v%v.%v.%v", segs[0], segs[1], segs[2]+1)
	println(ver)
	r = release{Version: ver}
	assert.Equal(t, true, r.IsApplicable())

	r = release{Version: ver + "-beta"}
	assert.Equal(t, false, r.IsApplicable())

	r = release{Version: ver + "+prerelease"}
	assert.Equal(t, false, r.IsApplicable())
}

func TestReleaseGetVersion(t *testing.T) {
	r := release{Version: "20.0.0"}
	assert.Equal(t, "20.0.0", r.GetVersion().String())
}

func TestReleaseForCurrentPlatform(t *testing.T) {
	thisSystem := runtime.GOOS + "-" + runtime.GOARCH

	r := release{Version: "20.0.0", Platforms: []platform{
		{Name: thisSystem},
		{Name: "other-system"},
		{Name: "bad-system"},
	}}
	assert.Equal(t, thisSystem, r.ForCurrentPlatform().Name)

	r = release{Version: "20.0.0", Platforms: []platform{
		{Name: "other-system"},
		{Name: "bad-system"},
	}}
	assert.Equal(t, "", r.ForCurrentPlatform().Name)
}

func TestReleasesListGet(t *testing.T) {
	var releases releasesList
	json.Unmarshal([]byte(`[{"version":"0.4.4"},{"version":"0.4.7"}]`), &releases)

	r := releases.Get("latest")
	assert.Equal(t, "0.4.7", r.Version)

	r = releases.Get("0.4.4")
	assert.Equal(t, "0.4.4", r.Version)

	r = releases.Get("0.0.0")
	assert.Equal(t, "", r.Version)
	assert.Equal(t, false, r.IsValid())
}
