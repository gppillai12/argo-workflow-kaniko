package config

import (
	errors "github.com/icelolly/go-errors"

	"github.com/spf13/viper"
)

type ClusterConfig struct {
	ClusterName       string          `json:"clusterName"`
	Version           string          `json:"version"`
	KubernetesVersion string          `json:"kubernetesVersion"`
	Infrastructure    *Infrastructure `json:"infrastructure"`
	Deployment        *Deployment     `json:"deployment"`
	Docker            *Docker         `json:"docker"`
}
type Pull struct {
	Registry string      `json:"registry"`
	Auth     *DockerAuth `json:"auth"`
	Images   []string    `json:"images"`
}
type Push struct {
	Registry string      `json:"registry"`
	Auth     *DockerAuth `json:"auth"`
}

type DockerAuth struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type Bundle struct {
	Path string `json:"path"`
	Name string `json:"name"`
}

type Docker struct {
	Pull   []*Pull  `json:"pull"`
	Build  []string `json:"build"`
	Push   *Push    `json:"push"`
	Bundle *Bundle  `json:"bundle"`
}

// Deployment is the specification for the cluster api components
type Deployment struct {
	Provider *DeploymentProvider `json:"provider"`
}

// DeploymentProvider is the Provider specification for the Deployment
type DeploymentProvider struct {
	Name      string      `json:"name"`
	NameSpace string      `json:"namespace"`
	Spec      interface{} `json:"spec"`
}

type Infrastructure struct {
	// The definition of the provider to use when creating infrastructure for the cluster.
	Provider *InfrastructureProvider `json:"provider"`
}

// InfrastructureProvider is the specification for the infrastructure provider.
type InfrastructureProvider struct {
	// The name of the provider to use.
	Name string `json:"name"`
	// The version kubernetes cluster
	KubernetesVersion string `json:"kubernetesVersion"`
	// The specification for the provider.
	Spec interface{} `json:"spec"`
}

// NewFromFile returns a new cluster config structure using the given file.
// Usually this will be used for loading config from a file.
func NewFromFile(filename string) (*ClusterConfig, error) {
	viper.SetConfigType("yaml")
	viper.SetConfigFile(filename)

	if err := viper.ReadInConfig(); err != nil {
		return nil, errors.Wrap(err, "failed to read config file")
	}

	var c ClusterConfig
	err := viper.Unmarshal(&c)
	if err != nil {
		return nil, err
	}

	return &c, nil
}
