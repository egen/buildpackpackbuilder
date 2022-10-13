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
	OutputPath        string
	VersionFolderPath string
	BuildFolderPath   string
}

var GITHUB_FORMAT string = "https://github.com/cloudfoundry/%s/archive/refs/tags/%s"
var GITHUB_GIT_FORMAT string = "https://github.com/cloudfoundry/%s.git"
var GITHUB_TAR string = "v%s.tar.gz"
var WD_OUTPUT string = ""

func (buildpack *BuildPack) CreateBuildPackDirectory() error {
	workdir, errdir := os.Getwd()
	if errdir != nil {
		log.Fatalln("Could not get working directory!")
		return errdir
	}

	WD_OUTPUT = filepath.Join(workdir, "out")

	//Create primary output directory if it doesn't exist.
	if _, err := os.Stat(WD_OUTPUT); os.IsNotExist(err) {
		err = os.Mkdir(WD_OUTPUT, os.ModeDir)
		if err != nil {
			log.Println("Output already exists.")
		}
	}

	buildpack.FullFolderPath = filepath.Join(workdir, buildpack.Name)
	buildpack.VersionFolderPath = filepath.Join(buildpack.FullFolderPath, fmt.Sprintf("%s-%s", buildpack.Name, buildpack.Version))
	buildpack.BuildFolderPath = buildpack.VersionFolderPath
	buildpack.OutputPath = buildpack.BuildFolderPath

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

func (buildpack *BuildPack) downloadOfficialGitHubRepository() error {

	buildpack.BuildFolderPath = path.Join(buildpack.BuildFolderPath, buildpack.Name)
	buildpack.OutputPath = buildpack.BuildFolderPath

	if _, err := os.Stat(buildpack.VersionFolderPath); os.IsNotExist(err) {
		err = os.Mkdir(buildpack.VersionFolderPath, os.ModeDir)
		if err != nil {
			return err
		}

		log.Printf("Created Directory %s", buildpack.VersionFolderPath)
	}

	log.Printf("[%s] Cloning Repo", buildpack.Name)
	err := buildpack.runCommand("git", []string{"clone", "--depth=1", "--branch", fmt.Sprintf("v%s", buildpack.Version), fmt.Sprintf(GITHUB_GIT_FORMAT, buildpack.Name)}, buildpack.VersionFolderPath, nil)
	if err != nil {
		return err
	}

	log.Printf("[%s] Getting submodules", buildpack.Name)
	err = buildpack.runCommand("git", []string{"submodule", "update", "--init"}, buildpack.BuildFolderPath, nil)
	if err != nil {
		return err
	}

	log.Printf("[%s] Checking out version: Clone successful", buildpack.Name)
	return nil
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

	switch buildtype := buildpack.Build.Type; buildtype {
	case "packager":
		return buildpack.RunBuildPackPackager()
	case "oldpackager":
		return buildpack.RunOldBuildPackPackager()
	case "custom":
		return buildpack.RunCustomPackager()
	case "java":
		return buildpack.RunJavaBuildPackPackager()
	default:
		return buildpack.RunBuildPackPackager()
	}

	return nil
}

func (buildpack *BuildPack) FindPackagedBuildpack() ([]string, error) {
	files, err := os.ReadDir(buildpack.OutputPath)
	if err != nil {
		return nil, err
	}

	foundFiles := make([]string, 0)

	for _, file := range files {
		if file.IsDir() == false {
			fullpath := path.Join(buildpack.OutputPath, file.Name())
			if filepath.Ext(fullpath) == ".zip" {
				foundFiles = append(foundFiles, fullpath)
			}
		}
	}

	return foundFiles, nil
}

func (buildpack *BuildPack) MoveArtifactToOutputDirectory() error {
	foundPacks, err := buildpack.FindPackagedBuildpack()
	if err != nil {
		return err
	}
	for _, file := range foundPacks {
		filename := path.Base(file)
		output := path.Join(WD_OUTPUT, filename)
		err := os.Rename(file, output)
		if err != nil {
			return err
		}
	}
	return nil
}

func (buildpack *BuildPack) RunBuildPackPackager() error {
	if _, err := os.Stat(buildpack.BuildFolderPath); !os.IsNotExist(err) {
		log.Printf("[%s] Started building buildpack...", buildpack.Name)

		err = buildpack.runCommand("buildpack-packager", []string{"build", "-stack", buildpack.Stack, "--cached", fmt.Sprintf("%t", buildpack.Offline)}, buildpack.BuildFolderPath, nil)
		if err != nil {
			return err
		}
		return nil

		log.Printf("[%s] Completed building buildpack!", buildpack.Name)
	}
	return nil
}

func (buildpack *BuildPack) RunOldBuildPackPackager() error {
	if _, err := os.Stat(buildpack.VersionFolderPath); !os.IsNotExist(err) {
		log.Printf("[%s] Started building buildpack using OLD BuildPack Packager...", buildpack.Name)

		err := buildpack.runCommand_Ruby_Bundle_Install()
		if err != nil {
			return err
		}

		log.Printf("[%s] Installed Bundle...", buildpack.Name)

		cached := "--cached"
		if buildpack.Offline {
			cached = "--uncached"
		}

		log.Printf("[%s] Building package...", buildpack.Name)

		err = buildpack.runCommand("bundle", []string{"exec", "buildpack-packager", cached, fmt.Sprintf("--stack=%s", buildpack.Stack)}, buildpack.BuildFolderPath, []string{"BUNDLE_GEMFILE=cf.GemFile"})
		if err != nil {
			return err
		}
		return nil
		log.Printf("[%s] Completed building buildpack!", buildpack.Name)
	}
	return nil
}

func (buildpack *BuildPack) RunJavaBuildPackPackager() error {
	if _, err := os.Stat(buildpack.VersionFolderPath); !os.IsNotExist(err) {
		log.Printf("[%s] Started building Java buildpack...", buildpack.Name)

		err = buildpack.runCommand_Ruby_Bundle_Install()
		if err != nil {
			return err
		}

		err := buildpack.runCommand("bundle", []string{"exec", "rake", "clean", "package", fmt.Sprintf("OFFLINE=%t", buildpack.Offline)}, buildpack.BuildFolderPath, nil)
		if err != nil {
			return err
		}
		return nil

		log.Printf("[%s] Completed building buildpack!", buildpack.Name)
	}
	return nil
}

func (buildpack *BuildPack) runCommand_Ruby_Bundle_Install() error {

	err := buildpack.runCommand("bundle", []string{"install"}, buildpack.BuildFolderPath, []string{"BUNDLE_GEMFILE=cf.GemFile"})
	if err != nil {
		return err
	}
	return nil
}

func (buildpack *BuildPack) runCommand(command string, args []string, workingdirectory string, env []string) error {

	cmd := exec.Command(command, args...)
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, env...)
	if DEBUG_MODE {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}
	cmd.Dir = workingdirectory
	err := cmd.Run()
	if err != nil {
		return err
	}
	return nil
}

func (buildpack *BuildPack) RunCustomPackager() error {
	if _, err := os.Stat(buildpack.BuildFolderPath); !os.IsNotExist(err) {
		log.Printf("[%s] Started building buildpack using CUSTOM...", buildpack.Name)
		cmd := exec.Command(buildpack.Build.Exec.Cmd, buildpack.Build.Exec.Args...)
		cmd.Dir = buildpack.BuildFolderPath
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

		return nil
	}
	return nil
}

func (buildpack *BuildPack) GetResource() error {
	if buildpack.Official {

		if buildpack.Type == "tar" {
			filename, err := buildpack.downloadOfficialGithubFile()
			if err != nil {
				return err
			}
			err = buildpack.ExpandTarResources(filename)
			if err != nil {
				return err
			}
		} else if buildpack.Type == "git" {
			err := buildpack.downloadOfficialGitHubRepository()
			if err != nil {
				return err
			}
		} else {
			return fmt.Errorf("Invalid buildpack download type specified")
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
