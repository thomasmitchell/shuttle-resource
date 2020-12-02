package main

import (
	"encoding/json"
	"fmt"
	"os"
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
	Skip              bool `json:"skip"`
	Return            bool `json:"return"`
	Fail              bool `json:"fail"`
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

	mode := parseMode(cfg.Params)

	var metadata []map[string]string
	var payload *models.Payload

	if mode == modeSkip {
		writeOutput(
			cfg.Version,
			[]map[string]string{
				{
					"name":  "skipped",
					"value": "true",
				},
			},
		)
		return
	}

	drv, err := driver.New(cfg.Source)
	if err != nil {
		utils.Bail("Failed to initialize driver: %s", err)
	}

	payload, err = drv.Read(cfg.Version)
	if err != nil {
		utils.Bail("Error when reading from remote: %s", err)
	}

	switch mode {
	case modeGet:

	case modeReturn:
		payload.Done = true
		payload.Passed = !cfg.Params.Fail
		err = drv.Write(cfg.Version, *payload)
		if err != nil {
			utils.Bail("Error writing to version `%s': %s", cfg.Version, err)
		}

	case modeWait:
		for !payload.Done {
			time.Sleep(30 * time.Second)
			fmt.Fprintf(os.Stderr, "Resource not ready. Waiting...\n")
			payload, err = drv.Read(cfg.Version)
			if err != nil {
				utils.Bail("Error when reading from remote: %s", err)
			}
		}

		err := drv.Clean(cfg.Version)
		if err != nil {
			utils.Bail("Error when cleaning up remote: %s", err)
		}

		if !cfg.Params.ContinueOnFailure && !payload.Passed {
			utils.Bail("Remote job returned with failure!")
		}
	}

	writeOutput(cfg.Version, genMetadata(payload))
}

type modeT int

const (
	modeGet modeT = iota
	modeSkip
	modeReturn
	modeWait
)

func parseMode(p Params) modeT {
	modes := 0
	ret := modeGet

	if p.Skip {
		ret = modeSkip
		modes++
	}

	if p.Wait {
		ret = modeWait
		modes++
	}

	if p.Return {
		ret = modeReturn
		modes++
	}

	if modes > 1 {
		utils.Bail("Must only specify one of skip, return, or wait")
	}

	if p.Fail && !p.Return {
		utils.Bail("Cannot specify fail without return")
	}

	if p.ContinueOnFailure && !p.Wait {
		utils.Bail("Cannot specify continue_on_failure without wait")
	}

	return ret
}

func writeOutput(version models.Version, metadata []map[string]string) {
	enc := json.NewEncoder(os.Stdout)
	err := enc.Encode(&Output{
		Version:  version,
		Metadata: metadata,
	})
	if err != nil {
		utils.Bail("Could not encode output JSON: %s", err)
	}
}

func genMetadata(payload *models.Payload) []map[string]string {
	return []map[string]string{
		{
			"name":  "caller",
			"value": payload.Caller,
		},
	}
}
