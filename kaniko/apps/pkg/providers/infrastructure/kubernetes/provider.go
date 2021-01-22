package kubernetes

import (
	"path/filepath"
	"yala/pkg/config"
	"yala/pkg/providers"

	"github.com/icelolly/go-errors"
	"github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

type ClusterProvider struct {
	Name        string
	ClusterName string
	WorkingDir  string
	Spec        *InfrastructureSpec
}

// Apply creates all resources required by the provider.
func (p *ClusterProvider) Apply() error {
	// This return nil for existing cluster
	return nil
}

// Delete deletes all resources created by the provider.
func (p *ClusterProvider) Delete() error {
	// This return nil for existing cluster
	return nil
}

// GetKubeConfig return kubeconfig file location
func (p *ClusterProvider) GetKubeConfig() (string, error) {
	kubeconfig, err := homedir.Expand(p.Spec.KubeConfig)
	if err != nil {
		return "", errors.Wrap(err, "unable to resolve kubeconfig file path")
	}
	return filepath.Abs(kubeconfig)
}

// New creates a new instance of the ClusterProvider
func New(homeDir string, c *config.ClusterConfig) (*ClusterProvider, error) {

	if c == nil {
		return nil, errors.New("the config object needs to be initialized, got nil")
	}

	var spec InfrastructureSpec

	workingDir, err := providers.CreateProviderDir(homeDir, c.ClusterName, c.Infrastructure.Provider.Name)
	if err != nil {
		return nil, err
	}

	err = viper.UnmarshalKey("infrastructure.provider.spec", &spec)
	if err != nil {
		return nil, errors.Wrap(err, "could not unmarshal infrastructure provider spec to provider.InfrastructureSpec")
	}

	return &ClusterProvider{
		Name:        c.Infrastructure.Provider.Name,
		ClusterName: c.ClusterName,
		WorkingDir:  workingDir,
		Spec:        &spec,
	}, nil
}
