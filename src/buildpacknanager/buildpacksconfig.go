package buildpackmanager

import (
	"io/ioutil"

	"gopkg.in/yaml.v3"
)

type BuildPacksConfig struct {
	BuildPacks []BuildPack `yaml:"buildpacks"`
	Debug      bool        `yaml:"debug"`
}

var DEBUG_MODE bool = false

func (bpc *BuildPacksConfig) loadBuildpackData(filename string) error {
	file, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}
	err = yaml.Unmarshal(file, bpc)
	if err != nil {
		return err
	}

	DEBUG_MODE = bpc.Debug //Sets debug mode if enabled

	return nil
}
