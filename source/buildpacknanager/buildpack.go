package buildpackmanager

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
)

type GitLocation struct {
	URL string `yaml:"repo"`
}

type Build struct {
	Type string `yaml:"type"`
	Exec Exec   `yaml:"exec"`
}

type Exec struct {
	Cmd  string   `yaml:"cmd"`
	Args []string `yaml:"args,flow"`
}

type TarLocation struct {
	URL string `yaml:"url"`
}

type BuildPack struct {
	Name              string      `yaml:"name"`
	Version           string      `yaml:"version"`
	Stack             string      `yaml:"stack"`
	Official          bool        `yaml:"official"`
	Type              string      `yaml:"type"`
	Offline           bool        `yaml:"offline"`
	GitLocation       GitLocation `yaml:"git"`
	TarLocation       TarLocation `yaml:"tar"`
	Build             Build       `yaml:"build"`
	Skip              bool        `yaml:"skip"`
	FullFolderPath    string
	VersionFolderPath string
}

var GITHUB_FORMAT string = "https://github.com/cloudfoundry/%s/archive/refs/tags/%s"
var GITHUB_TAR string = "v%s.tar.gz"

func (buildpack *BuildPack) CreateBuildPackDirectory() error {
	workdir, errdir := os.Getwd()
	if errdir != nil {
		log.Fatalln("Could not get working directory!")
		return errdir
	}

	buildpack.FullFolderPath = filepath.Join(workdir, buildpack.Name)
	buildpack.VersionFolderPath = filepath.Join(buildpack.FullFolderPath, fmt.Sprintf("%s-%s", buildpack.Name, buildpack.Version))

	if _, err := os.Stat(buildpack.FullFolderPath); os.IsNotExist(err) {
		err = os.Mkdir(buildpack.FullFolderPath, os.ModeDir)
		if err != nil {
			return err
		}
		log.Printf("Created Directory %s", buildpack.FullFolderPath)
	}
	return nil
}

func (buildpack *BuildPack) downloadOfficialGithubFile() (string, error) {

	filename := fmt.Sprintf(GITHUB_TAR, buildpack.Version)
	downloadfile := path.Join(buildpack.FullFolderPath, filename)
	fullURLdownloadfile := fmt.Sprintf(GITHUB_FORMAT, buildpack.Name, filename)
	if file, err := os.Stat(downloadfile); !os.IsNotExist(err) {
		if file.Size() > 0 {
			log.Printf("[%s] File: %s already exists! Skipping Download", buildpack.Name, filename)
			return downloadfile, nil
		}
	}

	file, err := os.Create(downloadfile)
	if err != nil {
		return "", err
	}

	log.Printf("[%s] File: %s Starting Download...", buildpack.Name, filename)

	client := http.Client{
		CheckRedirect: func(r *http.Request, via []*http.Request) error {
			r.URL.Opaque = r.URL.Path
			return nil
		},
	}

	resp, err := client.Get(fullURLdownloadfile)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	size, err := io.Copy(file, resp.Body)
	if err != nil {
		return "", err
	}
	defer file.Close()

	log.Printf("[%s] File: %s Downloaded Size %d", buildpack.Name, filename, size)

	return downloadfile, nil
}

func (buildpack *BuildPack) downloadTarFile(url string) (string, error) {

	filename := fmt.Sprintf(GITHUB_TAR, buildpack.Version)
	downloadfile := path.Join(buildpack.FullFolderPath, filename)
	fullURLdownloadfile := url
	if _, err := os.Stat(downloadfile); !os.IsNotExist(err) {
		log.Printf("[%s] File: %s already exists! Skipping Download", buildpack.Name, filename)
		return downloadfile, nil
	}

	file, err := os.Create(downloadfile)
	if err != nil {
		return "", err
	}

	log.Printf("[%s] File: %s Starting Download...", buildpack.Name, filename)

	client := http.Client{
		CheckRedirect: func(r *http.Request, via []*http.Request) error {
			r.URL.Opaque = r.URL.Path
			return nil
		},
	}

	resp, err := client.Get(fullURLdownloadfile)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	size, err := io.Copy(file, resp.Body)
	if err != nil {
		return "", err
	}
	defer file.Close()

	log.Printf("[%s] File: %s Downloaded Size %d", buildpack.Name, filename, size)

	return downloadfile, nil
}

func (buildpack *BuildPack) BuildBuildPack() error {
	if buildpack.Build.Type == "packager" {
		return buildpack.RunBuildPackPackager()
	} else if buildpack.Build.Type == "custom" {
		return buildpack.RunCustomPackager()
	}

	return nil
}

func (buildpack *BuildPack) RunBuildPackPackager() error {
	if _, err := os.Stat(buildpack.VersionFolderPath); !os.IsNotExist(err) {
		log.Printf("[%s] Started building buildpack...", buildpack.Name)
		cmd := exec.Command("buildpack-packager", "build", "-stack", buildpack.Stack, "--cached", fmt.Sprintf("%t", buildpack.Offline))
		cmd.Dir = buildpack.VersionFolderPath
		err := cmd.Run()
		if err != nil {
			return err
		}
		log.Printf("[%s] Completed building buildpack!", buildpack.Name)
	}
	return nil
}

func (buildpack *BuildPack) RunCustomPackager() error {
	if _, err := os.Stat(buildpack.VersionFolderPath); !os.IsNotExist(err) {
		log.Printf("[%s] Started building buildpack...", buildpack.Name)
		cmd := exec.Command(buildpack.Build.Exec.Cmd, buildpack.Build.Exec.Args...)
		cmd.Dir = buildpack.VersionFolderPath
		cmd.Stderr = os.Stdout
		cmd.Stdout = os.Stdout
		err := cmd.Run()
		if err != nil {
			return err
		}
		log.Printf("[%s] Completed building buildpack!", buildpack.Name)
	}
	return nil
}

func (buildpack *BuildPack) ExpandTarResources(filename string) error {
	if _, err := os.Stat(buildpack.VersionFolderPath); os.IsNotExist(err) {
		log.Printf("[%s] Started Extraction of source %s", buildpack.Name, buildpack.VersionFolderPath)
		cmd := exec.Command("tar", "-zxvf", filename, "-C", buildpack.FullFolderPath)
		err := cmd.Run()
		if err != nil {
			return err
		}
		log.Printf("[%s] Completed Extraction", buildpack.Name)
	}
	return nil
}

func (buildpack *BuildPack) GetResource() error {
	if buildpack.Official {
		filename, err := buildpack.downloadOfficialGithubFile()
		if err != nil {
			return err
		}
		err = buildpack.ExpandTarResources(filename)
		if err != nil {
			return err
		}
	} else {
		if buildpack.Type == "tar" {
			if buildpack.TarLocation.URL == "" {
				return fmt.Errorf("tar URL was not specified")
			}
			filename, err := buildpack.downloadTarFile(buildpack.TarLocation.URL)
			if err != nil {
				return err
			}
			err = buildpack.ExpandTarResources(filename)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
