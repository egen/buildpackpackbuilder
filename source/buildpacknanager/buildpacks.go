package buildpackmanager

import (
	"io/ioutil"

	"gopkg.in/yaml.v3"
)

type BuildPacks struct {
	BuildPack []BuildPack `yaml:"buildpacks"`
}

func (buildpacks *BuildPacks) loadBuildpackData(filename string) error {
	file, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}
	err = yaml.Unmarshal(file, buildpacks)
	if err != nil {
		return err
	}

	return nil
}
