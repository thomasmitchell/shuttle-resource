package driver

import "github.com/thomasmitchell/shuttle-resource/models"

type Driver interface {
	Read(version models.Version) (*models.Payload, error)
	Write(version models.Version, payload models.Payload) error
	LatestVersion() (models.Version, error)
  Clean(version models.Version) error
}

func New(cfg models.Source) (Driver, error) {
	return newS3Driver(cfg)
}
