package main

import (
	"log"
	"os"

	buildpackmanager "github.com/starkandwayne/buildpackpackbuilder/buildpacknanager"
)

type Configuration struct {
	BuildpackConfig  string
	DockerfileConfig string
}

var configuration Configuration

func main() {
	configuration = Configuration{}
	if len(os.Args) == 1 {
		log.Fatal("No arguements provided for file to load.")
	}

	configuration.BuildpackConfig = os.Args[1]

	log.Printf("Loading file: %s", configuration.BuildpackConfig)

	manager := buildpackmanager.Manager{}
	err := manager.Load(configuration.BuildpackConfig)
	if err != nil {
		log.Fatal(err)
	}

	err = manager.Process()
	if err != nil {
		log.Fatal(err)
	}

	//fmt.Printf("Data: %#v", manager.BuildPacks)
}
