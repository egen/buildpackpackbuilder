package buildpackmanager

import (
	"log"
	"sync"
)

type Manager struct {
	BuildPacks *BuildPacks
}

var wg sync.WaitGroup

func (bpm *Manager) Load(filename string) error {
	bpm.BuildPacks = &BuildPacks{}
	err := bpm.BuildPacks.loadBuildpackData(filename)
	if err != nil {
		return err
	}

	return nil
}

func (bpm *Manager) Process() error {
	for _, b := range bpm.BuildPacks.BuildPack {
		wg.Add(1)
		go func(b BuildPack) {
			log.Printf("[%s] Starting...", b.Name)
			defer wg.Done()
			err := b.CreateBuildPackDirectory()
			if err != nil {
				log.Println(err)
				return
			}
			err = b.GetResource()
			if err != nil {
				log.Println(err)
				return
			}

			err = b.BuildBuildPack()
			if err != nil {
				log.Println(err)
				return
			}
		}(b)

	}

	wg.Wait()

	return nil
}
