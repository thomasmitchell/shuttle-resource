package models

import (
	"strconv"

	"github.com/thomasmitchell/shuttle-resource/utils"
)

type Payload struct {
	Caller string `json:"caller"`
	Done   bool   `json:"done"`
	Passed bool   `json:"passed"`
}

type Source struct {
	Bucket          string `json:"bucket"`
	AccessKeyID     string `json:"access_key_id"`
	SecretAccessKey string `json:"secret_access_key"`
	Region          string `json:"region"`
	Path            string `json:"path"`
	Endpoint        string `json:"endpoint"`
	TLSSkipVerify   bool   `json:"tls_skip_verify"`
}

type Version struct {
	Number string `json:"number"`
}

func NewVersion(number string) Version {
	return Version{Number: number}
}

func (v Version) LessThan(v2 Version) bool {
	return v.Uint64() < v2.Uint64()
}

func (v Version) Increment() Version {
	return Version{Number: strconv.FormatUint(v.Uint64()+1, 10)}
}

func (v Version) Decrement() Version {
	return Version{Number: strconv.FormatUint(v.Uint64()-1, 10)}
}

func (v Version) Uint64() uint64 {
	ret, err := strconv.ParseUint(v.Number, 10, 64)
	if err != nil {
		utils.Bail("Unsupported number format for version `%s': %s", v.Number, err)
	}

	return ret
}

func (v Version) String() string {
	return v.Number
}
