package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/thomasmitchell/shuttle-resource/driver"
	"github.com/thomasmitchell/shuttle-resource/models"
	"github.com/thomasmitchell/shuttle-resource/utils"
)

type Config struct {
	Source models.Source `json:"source"`
	Params Params        `json:"params"`
}

type Params struct {
	Return string `json:"return"`
	Fail   bool   `json:"fail"`
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

	var (
		pay *models.Payload
		ver models.Version
	)

	if cfg.Params.Return != "" {
		pay, ver, err = getReturnValue(drv, cfg.Params.Return, cfg.Params.Fail)
	} else {
		pay, ver, err = getNewRequest(drv)
	}

	err = drv.Write(ver, *pay)
	if err != nil {
		utils.Bail("Error writing to version `%s': %s", ver.Number, err)
	}

	//Get this resource's latest version in case we bumped it
	version, err := drv.LatestVersion()
	if err != nil {
		utils.Bail("Erred getting our own version")
	}

	output := Output{
		Version:  version,
		Metadata: []map[string]string{},
	}
	enc := json.NewEncoder(os.Stdout)
	err = enc.Encode(&output)
	if err != nil {
		utils.Bail("Could not encode output JSON: %s", err)
	}
}

func readVersionFromFile(path string) (models.Version, error) {
	ret := models.Version{}
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			if _, statErr := os.Stat(filepath.Dir(path)); statErr != nil {
				if os.IsNotExist(statErr) {
					//Resource input isn't there
					return ret,
						fmt.Errorf(
							"Missing resource input `%s'",
							filepath.Base(filepath.Dir(path)),
						)
				}
			}

			//Path file isn't there, which can just mean there's nothing to return
			return ret, nil
		}

		//Actual IO error
		return ret, fmt.Errorf("Error opening file `%s': %s", path, err)
	}
	defer f.Close()

	dec := json.NewDecoder(f)
	err = dec.Decode(&ret)
	if err != nil {
		return ret, fmt.Errorf("Error decoding JSON from file `%s': %s", path, err)
	}

	return ret, nil
}

func getReturnValue(drv driver.Driver, resource string, fail bool) (*models.Payload, models.Version, error) {
	pay := &models.Payload{}
	ver := models.Version{}
	ver, err := readVersionFromFile(
		filepath.Join(os.Args[1], resource, "version"),
	)
	if err != nil {
		return pay, ver, fmt.Errorf(
			"Could not retrieve version `%s' from resource: %s",
			ver,
			err,
		)
	}

	pay, err = drv.Read(ver)
	if err != nil {
		return pay, ver, fmt.Errorf(
			"Error when retrieving remote version `%s': %s",
			ver.Number,
			err,
		)
	}

	pay.Done = true
	pay.Passed = !fail
	return pay, ver, err
}

func getNewRequest(drv driver.Driver) (*models.Payload, models.Version, error) {
	ver := models.Version{}
	remoteLatest, err := drv.LatestVersion()
	if err != nil {
		return nil, ver, fmt.Errorf("Error fetching latest version: %s", err)
	}

	ver = remoteLatest.Increment()
	pay := &models.Payload{
		Caller: utils.CallerName(),
	}

	return pay, ver, nil
}
