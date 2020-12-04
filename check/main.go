package main

import (
	"encoding/json"
	"os"
	"sort"

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

	drv, err := driver.New(cfg.Source)
	versions, err := drv.Versions()
	if err != nil {
		utils.Bail("Error when reading from remote: %s", err)
	}

	if cfg.Source.SeedVersion {
		versions = append(versions, models.VersionFromInt(1))
	}

	output := versions.Since(cfg.Version)
	if cfg.Version.Int64() == 0 && len(output) > 0 {
		//only return latest if concourse doesn't already have a version
		sort.Sort(output)
		output = output[len(output)-1:]
	}

	enc := json.NewEncoder(os.Stdout)
	err = enc.Encode(output)
	if err != nil {
		utils.Bail("Error encoding output value: %s", err)
	}
}
