package main

import (
	"encoding/json"
	"os"

	"github.com/thomasmitchell/shuttle-resource/driver"
	"github.com/thomasmitchell/shuttle-resource/models"
	"github.com/thomasmitchell/shuttle-resource/utils"
)

type Config struct {
	Source  models.Source  `json:"source"`
	Version models.Version `json:"version"`
}

func main() {
	dec := json.NewDecoder(os.Stdin)
	cfg := &Config{}
	err := dec.Decode(&cfg)
	if err != nil {
		utils.Bail("Failed to decode input JSON: %s", err)
	}

	if cfg.Version.Number == "" {
		cfg.Version.Number = "0"
	}

	drv, err := driver.New(cfg.Source)
	remoteVersion, err := drv.LatestVersion()
	if err != nil {
		utils.Bail("Error when reading from remote: %s", err)
	}

	output := []models.Version{}
	for cfg.Version.LessThan(remoteVersion) {
		output = append(output, remoteVersion)
		remoteVersion = remoteVersion.Decrement()
	}

	enc := json.NewEncoder(os.Stdout)
	err = enc.Encode(output)
	if err != nil {
		utils.Bail("Error encoding output value: %s", err)
	}
}
