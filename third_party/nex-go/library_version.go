package nex

import (
	"fmt"
	"strings"

	"golang.org/x/mod/semver"
)


type LibraryVersion struct {
	Major             int
	Minor             int
	Patch             int
	GameSpecificPatch string
	semver            string
}


func (lv *LibraryVersion) Copy() *LibraryVersion {
	return &LibraryVersion{
		Major:             lv.Major,
		Minor:             lv.Minor,
		Patch:             lv.Patch,
		GameSpecificPatch: lv.GameSpecificPatch,
		semver:            fmt.Sprintf("v%d.%d.%d", lv.Major, lv.Minor, lv.Patch),
	}
}

func (lv *LibraryVersion) semverCompare(compare string) int {
	if !strings.HasPrefix(compare, "v") {

		var b strings.Builder

		b.WriteString("v")
		b.WriteString(compare)

		compare = b.String()
	}

	if !semver.IsValid(compare) {

		return 0
	}

	return semver.Compare(lv.semver, compare)
}


func (lv *LibraryVersion) GreaterOrEqual(compare string) bool {
	return lv.semverCompare(compare) != -1
}


func (lv *LibraryVersion) LessOrEqual(compare string) bool {
	return lv.semverCompare(compare) != 1
}


func NewPatchedLibraryVersion(major, minor, patch int, gameSpecificPatch string) *LibraryVersion {
	return &LibraryVersion{
		Major:             major,
		Minor:             minor,
		Patch:             patch,
		GameSpecificPatch: gameSpecificPatch,
		semver:            fmt.Sprintf("v%d.%d.%d", major, minor, patch),
	}
}


func NewLibraryVersion(major, minor, patch int) *LibraryVersion {
	return &LibraryVersion{
		Major:  major,
		Minor:  minor,
		Patch:  patch,
		semver: fmt.Sprintf("v%d.%d.%d", major, minor, patch),
	}
}


type LibraryVersions struct {
	Main         *LibraryVersion
	DataStore    *LibraryVersion
	MatchMaking  *LibraryVersion
	Ranking      *LibraryVersion
	Ranking2     *LibraryVersion
	Messaging    *LibraryVersion
	Utility      *LibraryVersion
	NATTraversal *LibraryVersion
}


func (lvs *LibraryVersions) SetDefault(version *LibraryVersion) {
	lvs.Main = version
	lvs.DataStore = version.Copy()
	lvs.MatchMaking = version.Copy()
	lvs.Ranking = version.Copy()
	lvs.Ranking2 = version.Copy()
	lvs.Messaging = version.Copy()
	lvs.Utility = version.Copy()
	lvs.NATTraversal = version.Copy()
}


func NewLibraryVersions() *LibraryVersions {
	return &LibraryVersions{}
}
