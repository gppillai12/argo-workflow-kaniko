package gitsource

import (
	"os"
	"path/filepath"
	"strings"
	"yala/pkg/command"

	errors "github.com/icelolly/go-errors"
)

// Data is a Source implementation for Git that only includes user provided values
type Data struct {
	URL      string `json:"url"`
	Name     string `json:"name"`
	Filename string `json:"filename"`
	Ref      string `json:"tag"`
	Force    bool   `json:"force"`
}

// GitSource is a Source implementation for Git
type GitSource struct {
	Data       Data
	WorkingDir string
	Runner     command.CmdRunner
}

// Name returns a reasonable filename for the source
func (git *GitSource) Name() (string, error) {
	if git.Data.Name == "" {
		return "", errors.New("name not initialized").WithField("data", git.Data)
	}
	if git.Data.Ref == "" {
		return "", errors.New("ref not initialized").WithField("data", git.Data)
	}
	return git.Data.Name, nil
}

// Name returns a reasonable filename for the source
func (git *GitSource) GetValuesFile() (string, error) {
	return "", nil
}

// Get downloads from a given GitSource and returns the directory data was pulled into
func (git *GitSource) Get() (string, error) {
	_, err := os.Open(filepath.Join(git.WorkingDir, git.Data.Name))
	if err == nil {
		if git.Data.Force {
			err = os.RemoveAll(filepath.Join(git.WorkingDir, git.Data.Name))
			if err != nil {
				return "", errors.Wrap(err, "could not clean up existing files")
			}
		} else {
			_, err = os.Open(filepath.Join(git.WorkingDir, git.Data.Name, git.Data.Filename))
			if err != nil {
				return "", errors.Wrap(err, "could not open existing files")
			}
			return filepath.Join(git.WorkingDir, git.Data.Name, git.Data.Filename), nil
		}
	}
	err = git.Runner.Run("git", "clone", "--single-branch", "--branch", git.Data.Ref, git.Data.URL, filepath.Join(git.WorkingDir, git.Data.Name))
	if err != nil {
		return "", err
	}
	return filepath.Join(git.WorkingDir, git.Data.Name, git.Data.Filename), nil
}

// New returns a GitSource given data and a commandRunner
func New(data Data, cmdRunner command.CmdRunner) (*GitSource, error) {
	workingDir, err := cmdRunner.Output("pwd")
	if err != nil {
		return nil, err
	}
	if data.URL == "" {
		return nil, errors.New("url not initialized").WithField("data", data)
	}
	if data.Name == "" {
		return nil, errors.New("name not initialized").WithField("data", data)
	}
	if data.Ref == "" {
		data.Ref = "master"
	}
	return &GitSource{
		Data:       data,
		WorkingDir: strings.TrimSpace(string(workingDir)),
		Runner:     cmdRunner,
	}, nil
}

// NewFromMap returns a GitSource given data in a map and a commandRunner. Does not typecheck!
func NewFromMap(data map[string]interface{}, cmdRunner command.CmdRunner, dir string) (*GitSource, error) {
	var gitData Data
	var ok bool
	gitData.URL, ok = data["url"].(string)
	if !ok {
		return nil, errors.New("url is not of type string").WithField("data", data)
	}
	gitData.Name, ok = data["name"].(string)
	if !ok {
		return nil, errors.New("name is not of type string").WithField("data", data)
	}
	gitData.Filename = ""
	gitData.Ref = "master"
	gitData.Force = false
	if filename, ok := data["filename"]; ok {
		gitData.Filename, ok = filename.(string)
		if !ok {
			return nil, errors.New("filename is not of type string").WithField("data", data)
		}
	}
	if ref, ok := data["ref"]; ok {
		gitData.Ref, ok = ref.(string)
		if !ok {
			return nil, errors.New("ref is not of type string").WithField("data", data)
		}
	}
	if force, ok := data["force"]; ok {
		gitData.Force, ok = force.(bool)
		if !ok {
			return nil, errors.New("force is not of type bool").WithField("data", data)
		}
	}
	gitSource, err := New(gitData, cmdRunner)
	if err != nil {
		return nil, errors.Wrap(err, "failed to initialize a gitsource")
	}
	gitSource.WorkingDir = dir
	return gitSource, nil
}
