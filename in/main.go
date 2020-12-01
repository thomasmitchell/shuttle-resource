package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/thomasmitchell/shuttle-resource/driver"
	"github.com/thomasmitchell/shuttle-resource/models"
	"github.com/thomasmitchell/shuttle-resource/utils"
)

type Config struct {
	Source  models.Source  `json:"source"`
	Version models.Version `json:"version"`
	Params  Params         `json:"params"`
}

type Params struct {
	Wait              bool `json:"wait"`
	ContinueOnFailure bool `json:"continue_on_failure"`
}

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

	payload, err := drv.Read(cfg.Version)
	if err != nil {
		utils.Bail("Error when reading from remote: %s", err)
	}

	if cfg.Params.Wait {
		for !payload.Done {
			time.Sleep(30 * time.Second)
			fmt.Fprintf(os.Stderr, "Resource not ready. Waiting...")
			payload, err = drv.Read(cfg.Version)
			if err != nil {
				utils.Bail("Error when reading from remote: %s", err)
			}
		}

		err := drv.Clean(cfg.Version)
		if err != nil {
			utils.Bail("Error when cleaning up remote: %s", err)
		}
	}

	if !cfg.Params.ContinueOnFailure && !payload.Passed {
		utils.Bail("Remote job returned with failure!")
	}

	writeJSON(filepath.Join(os.Args[1], "version"), &cfg.Version)

	enc := json.NewEncoder(os.Stdout)
	err = enc.Encode(&Output{
		Version: cfg.Version,
		Metadata: []map[string]string{
			{"name": "caller", "value": payload.Caller},
		},
	})
	if err != nil {
		utils.Bail("Could not encode output JSON: %s", err)
	}
}

func writeJSON(path string, obj interface{}) {
	f, err := os.Create(path)
	if err != nil {
		utils.Bail("Could not open up file at `%s': %s", err)
	}

	defer f.Close()

	enc := json.NewEncoder(f)
	err = enc.Encode(&obj)
	if err != nil {
		utils.Bail("Could not encode JSON to file: %s")
	}
}
