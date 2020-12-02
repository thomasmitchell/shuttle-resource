package models

import (
	"encoding/json"
	"sort"
	"strconv"
	"time"
)

type Version int64

func NewVersion() Version {
	return Version(time.Now().UnixNano())
}

func VersionFromInt(i int64) Version {
	return Version(i)
}

func VersionFromString(v string) (Version, error) {
	i, err := strconv.ParseInt(v, 10, 64)
	return Version(i), err
}

func (v Version) LessThan(v2 Version) bool { return v < v2 }
func (v Version) Increment() Version       { return v + 1 }
func (v Version) Int64() int64             { return int64(v) }
func (v Version) String() string {
	return strconv.FormatInt(int64(v), 10)
}

type versionJSON struct {
	Number string `json:"number"`
}

func (v *Version) MarshalJSON() ([]byte, error) {
	return json.Marshal(&versionJSON{Number: v.String()})
}

func (v *Version) UnmarshalJSON(in []byte) error {
	vJSON := versionJSON{}
	err := json.Unmarshal(in, &vJSON)
	if err != nil {
		return err
	}

	*v, err = VersionFromString(vJSON.Number)
	return err
}

type VersionList []Version

func (v VersionList) Len() int           { return len(v) }
func (v VersionList) Swap(i, j int)      { v[i], v[j] = v[j], v[i] }
func (v VersionList) Less(i, j int) bool { return v[i].LessThan(v[j]) }

func (v VersionList) Since(ver Version) VersionList {
	sort.Sort(v)
	idx := sort.Search(len(v), func(i int) bool { return ver.LessThan(v[i]) })
	return v[idx:]
}
