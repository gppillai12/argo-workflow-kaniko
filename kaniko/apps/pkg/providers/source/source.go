package source

import (
	"yala/pkg/command"
	filesource "yala/pkg/providers/source/file"
	gitsource "yala/pkg/providers/source/git"

	errors "github.com/icelolly/go-errors"
)

// Source is an interface for source locations
type Source interface {
	Get() (string, error)
	Name() (string, error)
	GetValuesFile() (string, error)
}

// New returns a Source given data
func NewSource(data interface{}, r command.CmdRunner) (Source, error) {
	if sourceData, ok := data.(gitsource.Data); ok {
		return gitsource.New(sourceData, r)
	} else if sourceData, ok := data.(filesource.Data); ok {
		return filesource.New(sourceData, r)
	}

	return nil, errors.New("source is not of any implemented type").WithFields("data", data)
}

// NewFromMap returns a source given data in a map
func NewFromMap(data map[string]interface{}, r command.CmdRunner, dir string) (Source, error) {
	if MapContainsSet(data, []string{"url", "name"}) {
		return gitsource.NewFromMap(data, r, dir)
	} else if MapContainsSet(data, []string{"path", "valuesFile"}) {
		return filesource.NewFromMap(data, r, dir)
	}

	return nil, errors.New("source is not of any implemented type").WithFields("data", data)
}

// MapContainsSet returns a boolean of whether a given map contains all of the elements in a given set
func MapContainsSet(data map[string]interface{}, set []string) bool {
	for _, item := range set {
		if _, ok := data[item]; !ok {
			return false
		}
	}
	return true
}
