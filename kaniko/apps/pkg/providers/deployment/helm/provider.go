package helm

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"yala/pkg/command"
	"yala/pkg/command/runner"
	"yala/pkg/config"
	"yala/pkg/providers"
	"yala/pkg/providers/source"

	errors "github.com/icelolly/go-errors"
	log "github.com/sirupsen/logrus"
)

// ClusterProvider implements the Provider interface.
type ClusterProvider struct {
	Name        string
	NameSpace   string
	ClusterName string
	WorkingDir  string
	KubeConfig  string
	Cmd         command.CmdRunner
	Spec        *DeploymentSpec
}

// Apply creates all resources required by the provider.
func (p *ClusterProvider) Apply() error {
	prov, err := p.Cmd.BinaryExists("helm")
	if err != nil {
		return errors.Wrap(err, "please install helm before running the command")
	}
	if !prov {
		return errors.New("please install helm before running the command")
	}
	charts := p.Spec.Charts
	for _, chart := range charts {
		chartSource, err := source.NewFromMap(chart, p.Cmd, filepath.Join(p.WorkingDir, "source"))
		if err != nil {
			return errors.Wrap(err, "failed to initialize a source for chart").WithField("chart", chart)
		}
		dir, err := chartSource.Get()
		if err != nil {
			return errors.Wrap(err, "failed to retrieve chart")
		}

		valuesFile, err := chartSource.GetValuesFile()
		if err != nil {
			return errors.Wrap(err, "unable to read values file")
		}

		name, err := chartSource.Name()
		if err != nil {
			return errors.Wrap(err, "could not get name for chart").WithField("chart", chart)
		}

		rendered, err := p.Cmd.Output("helm", "template", name, dir, "--namespace", p.NameSpace, "-f", valuesFile)
		if err != nil {
			return errors.Wrap(err, "could not template chart").WithField("directory", dir)
		}

		err = ioutil.WriteFile(filepath.Join(p.WorkingDir, "rendered", name), rendered, 0644)
		if err != nil {
			return errors.Wrap(err, "could not write file")
		}

		err = p.Cmd.Run("helm", "upgrade", "--install", "--namespace", p.NameSpace, "--create-namespace", name, "--kubeconfig", p.KubeConfig, dir, "-f", valuesFile)
		if err != nil {
			return errors.Wrap(err, "unable to deploy helm chart")
		}

	}
	return nil
}

// Delete deletes all resources created by the provider.
func (p *ClusterProvider) Delete() error {
	prov, err := p.Cmd.BinaryExists("helm")
	if err != nil {
		return errors.Wrap(err, "please install helm before running the command")
	}
	if !prov {
		return errors.New("please install helm before running the command")
	}
	charts := p.Spec.Charts
	for _, chart := range charts {
		chartSource, err := source.NewFromMap(chart, p.Cmd, filepath.Join(p.WorkingDir, "source"))
		if err != nil {
			return errors.Wrap(err, "failed to initialize a source for chart").WithField("chart", chart)
		}

		name, err := chartSource.Name()
		if err != nil {
			return errors.Wrap(err, "could not get name for chart").WithField("chart", chart)
		}

		p.Cmd.Run("helm", "uninstall", "--namespace", p.NameSpace, name, "--kubeconfig", p.KubeConfig)

	}
	return nil
}

// New creates a new instance of the ClusterProvider
func New(homeDir string, kubeConfig string, c *config.ClusterConfig, logger *log.Logger) (*ClusterProvider, error) {

	cmdRunner, err := runner.New(logger.Writer(), logger.Writer(), homeDir)

	if err != nil {
		log.WithError(err).WithFields(log.Fields{
			"workingDir": homeDir,
		}).Fatal("Failed to create command runner")
	}
	if c == nil {
		return nil, errors.New("invalid config")
	}
	if homeDir == "" {
		return nil, errors.New("invalid home directory")
	}

	spec, err := NewDeploymentSpec(c.Deployment.Provider.Spec)

	if err != nil {
		return nil, errors.Wrap(err, "could not parse deployment spec")
	}

	workingDir, err := providers.CreateProviderDir(homeDir, c.ClusterName, c.Deployment.Provider.Name)
	if err != nil {
		return nil, err
	}

	err = os.MkdirAll(filepath.Join(workingDir, "rendered"), 0750)
	if err != nil {
		return nil, err
	}

	err = os.MkdirAll(filepath.Join(workingDir, "source"), 0750)
	if err != nil {
		return nil, err
	}

	return &ClusterProvider{
		Name:        c.Infrastructure.Provider.Name,
		NameSpace:   c.Deployment.Provider.NameSpace,
		ClusterName: c.ClusterName,
		WorkingDir:  workingDir,
		KubeConfig:  kubeConfig,
		Cmd:         cmdRunner,
		Spec:        spec,
	}, nil
}
