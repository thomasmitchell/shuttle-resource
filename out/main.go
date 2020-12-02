package main

import (
	"encoding/json"
	"os"

	"github.com/thomasmitchell/shuttle-resource/driver"
	"github.com/thomasmitchell/shuttle-resource/models"
	"github.com/thomasmitchell/shuttle-resource/utils"
)

type Config struct {
	Source models.Source `json:"source"`
	Params Params        `json:"params"`
}

type Params struct{}

type Output struct {
	Version  models.Version      `json:"version"`
	Metadata []map[string]string `json:"metadata"`
}

func main() {
	dec := json.NewDecoder(os.Stdin)
	cfg := &Config{}
	err := dec.Decode(&cfg)
	if err != nil {
		utils.Bail("Failed to decode input JSON: %s", err)
	}

	drv, err := driver.New(cfg.Source)
	if err != nil {
		utils.Bail("Failed to initialize driver: %s", err)
	}

	ver := models.NewVersion()

	err = drv.Write(
		ver,
		models.Payload{Caller: utils.CallerName()},
	)
	if err != nil {
		utils.Bail("Error writing to version `%s': %s", ver, err)
	}

	output := Output{
		Version:  ver,
		Metadata: []map[string]string{},
	}

	enc := json.NewEncoder(os.Stdout)
	err = enc.Encode(&output)
	if err != nil {
		utils.Bail("Could not encode output JSON: %s", err)
	}
}
