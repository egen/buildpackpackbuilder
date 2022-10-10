package buildpackmanager

import (
	"log"
	"sync"
)

type Manager struct {
	BuildPacksConfig *BuildPacksConfig
}

var wg sync.WaitGroup

func (bpm *Manager) Load(filename string) error {
	bpm.BuildPacksConfig = &BuildPacksConfig{}
	err := bpm.BuildPacksConfig.loadBuildpackData(filename)
	if err != nil {
		return err
	}

	return nil
}

func (bpm *Manager) Process() error {
	for _, b := range bpm.BuildPacksConfig.BuildPacks {
		wg.Add(1)
		go func(b BuildPack) {
			defer wg.Done()
			if b.Skip {
				log.Printf("[%s] Skipping...", b.Name)
				return
			}
			log.Printf("[%s] Starting...", b.Name)
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

			err = b.MoveArtifactToOutputDirectory()
			if err != nil {
				log.Println(err)
				return
			}
		}(b)

	}

	wg.Wait()

	return nil
}
