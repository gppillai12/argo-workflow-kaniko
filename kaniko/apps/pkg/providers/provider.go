package providers

import (
	"os"
	"path/filepath"

	errors "github.com/icelolly/go-errors"
)

// CreateProviderDir creates the provider working directory in the Yala home directory.
// This should be used in the New method of each provider.
func CreateProviderDir(homeDir, clusterName, providerName string) (string, error) {
	if homeDir == "" || clusterName == "" || providerName == "" {
		return "", errors.New("clusterName and providerName must be set")
	}

	dir := filepath.Join(
		homeDir,
		clusterName,
		"provider",
		providerName,
	)

	return dir, os.MkdirAll(dir, 0750)

}
