package filesource

import (
	"path/filepath"
	"strings"
	"yala/pkg/command"
	"yala/pkg/file"

	errors "github.com/icelolly/go-errors"
	"github.com/mitchellh/go-homedir"
)

type Data struct {
	Name       string `json:"name"`
	Path       string `json:"path"`
	ValuesFile string `json:"valuesFile"`
}

type FileSource struct {
	Data       Data
	WorkingDir string
}

// Name returns a reasonable filename for the source
func (fileSource *FileSource) Name() (string, error) {
	if fileSource.Data.Name == "" {
		return "", errors.New("name not initialized").WithField("data", fileSource.Data)
	}

	return fileSource.Data.Name, nil
}

// Name returns a reasonable filename for the source
func (fileSource *FileSource) GetValuesFile() (string, error) {
	if fileSource.Data.ValuesFile == "" {
		return "", errors.New("name not initialized").WithField("data", fileSource.Data)
	}
	absPath, err := filepath.Abs(fileSource.Data.ValuesFile)
	if err != nil {
		return "", errors.Wrap(err, "unable to resolve file path")
	}
	err = file.CopyFile(absPath, filepath.Join(fileSource.WorkingDir, fileSource.Data.Name, ".values.yaml"))
	if err != nil {
		return "", errors.Wrap(err, "unable to copy valuesFile")
	}
	return filepath.Join(fileSource.WorkingDir, fileSource.Data.Name, ".values.yaml"), nil
}

// Get function copy directory for local path
func (fileSource *FileSource) Get() (string, error) {
	err := file.CopyDir(fileSource.Data.Path, filepath.Join(fileSource.WorkingDir, fileSource.Data.Name))
	if err != nil {
		return "", errors.Wrap(err, "unable to copy terraform module")
	}
	return filepath.Join(fileSource.WorkingDir, fileSource.Data.Name), nil
}

// New returns a GitSource given data and a commandRunner
func New(data Data, cmdRunner command.CmdRunner) (*FileSource, error) {
	workingDir, err := cmdRunner.Output("pwd")
	if err != nil {
		return nil, err
	}
	if data.Path == "" {
		return nil, errors.New("path not initialized").WithField("data", data)
	}
	if data.Name == "" {
		return nil, errors.New("name not initialized").WithField("data", data)
	}
	if data.ValuesFile == "" {
		return nil, errors.New("valuesFile not initialized").WithField("data", data)
	}

	return &FileSource{
		Data:       data,
		WorkingDir: strings.TrimSpace(string(workingDir)),
	}, nil
}

// NewFromMap returns a GitSource given data in a map and a commandRunner. Does not typecheck!
func NewFromMap(data map[string]interface{}, cmdRunner command.CmdRunner, dir string) (*FileSource, error) {
	var fileData Data
	var ok bool
	path, ok := data["path"].(string)
	if !ok {
		return nil, errors.New("path is not of type string").WithField("data", data)
	}

	fileData.Path, _ = homedir.Expand(path)

	fileData.Name, ok = data["name"].(string)
	if !ok {
		return nil, errors.New("name is not of type string").WithField("data", data)
	}

	valuesFile, ok := data["valuesFile"].(string)
	if !ok {
		return nil, errors.New("valuesFile is not of type string").WithField("data", data)
	}

	fileData.ValuesFile, _ = homedir.Expand(valuesFile)

	fileSource, err := New(fileData, cmdRunner)
	if err != nil {
		return nil, errors.Wrap(err, "failed to initialize a filesource")
	}
	fileSource.WorkingDir = dir
	return fileSource, nil
}
