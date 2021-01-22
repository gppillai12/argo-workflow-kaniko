package aws

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"text/template"
	"yala/pkg/command"
	"yala/pkg/command/runner"
	"yala/pkg/config"
	"yala/pkg/file"
	"yala/pkg/providers"

	errors "github.com/icelolly/go-errors"
	log "github.com/sirupsen/logrus"

	"github.com/spf13/viper"
)

type ClusterProvider struct {
	Name              string
	ClusterName       string
	KubernetesVersion string
	WorkingDir        string
	Spec              *InfrastructureSpec
	Cmd               command.CmdRunner
}

func (p *ClusterProvider) copyModule() error {
	err := file.CopyDir(p.Spec.ModuleSource.Path, filepath.Join(p.WorkingDir, p.Spec.ModuleSource.Name))
	if err != nil {
		return errors.Wrap(err, "unable to copy terraform module")
	}
	return nil
}

// Renders the terraform.tfvars file for the current provider spec.
func (p *ClusterProvider) renderVarsFile(writer io.Writer) error {
	t, err := template.New("terraform.tfvars").Parse(tfvarsTemplate)
	if err != nil {
		return errors.Wrap(err, "parse tfvars template failed")
	}
	err = t.Execute(writer, p)

	if err != nil {
		return errors.Wrap(err, "render tfvars failed")
	}
	return nil
}

func (p *ClusterProvider) manageVPC() error {
	if p.Spec.VpcID != "" {
		err := p.enableTFFile("vpc_existing.tf")
		if err != nil {
			return errors.Wrap(err, "failed to enable vpc_existing.tf")
		}
		err = p.disableTFFile("vpc_new.tf")
		if err != nil {
			return errors.Wrap(err, "failed to disable vpc_new.tf")
		}
	} else {
		err := p.enableTFFile("vpc_new.tf")
		if err != nil {
			return errors.Wrap(err, "failed to enable vpc_new.tf")
		}
		err = p.disableTFFile("vpc_existing.tf")
		if err != nil {
			return errors.Wrap(err, "failed to disable vpc_existing.tf")
		}
	}
	return nil
}

func (p *ClusterProvider) enableTFFile(filename string) error {
	modulesource := filepath.Join(p.WorkingDir, p.Spec.ModuleSource.Name, filename)

	if file.Exists(modulesource) {
		return nil
	}

	return os.Rename(modulesource+".disabled", modulesource)
}

func (p *ClusterProvider) disableTFFile(filename string) error {
	modulesource := filepath.Join(p.WorkingDir, p.Spec.ModuleSource.Name, filename)

	if file.Exists(modulesource + ".disabled") {
		return nil
	}

	return os.Rename(modulesource, modulesource+".disabled")
}

func (p *ClusterProvider) runModule() error {
	prov, err := p.Cmd.BinaryExists("terraform")
	if err != nil {
		return errors.Wrap(err, "please install Terraform before running the command")
	}
	if !prov {
		return errors.New("please install Terraform before running the command")
	}

	p.Cmd.SetDirectory(filepath.Join(p.WorkingDir, p.Spec.ModuleSource.Name))
	err = p.Cmd.Run("terraform", "init")
	if err != nil {
		return errors.Wrap(err, "run terraform init command failed")
	}

	err = p.Cmd.Run("terraform", "plan")
	if err != nil {
		return errors.Wrap(err, "run terraform plan command failed")
	}

	err = p.Cmd.Run("terraform", "apply", "-auto-approve", "-input=false")
	if err != nil {
		return errors.Wrap(err, "run terraform apply command failed")
	}

	return nil
}

func (p *ClusterProvider) Apply() error {
	err := p.copyModule()
	if err != nil {
		return errors.Wrap(err, "copy terraform module failed")
	}

	err = p.manageVPC()

	if err != nil {
		return errors.Wrap(err, "error occurred while handeling VPC")
	}

	filename := filepath.Join(p.WorkingDir, p.Spec.ModuleSource.Name, "terraform.tfvars")
	// If terraform.tfvars exist delete it, os.OpenFile somehow corrupting this file if we re-run yala apply with cluster config change.
	if file.Exists(filename) {
		err := os.Remove(filename)
		if err != nil {
			return errors.Wrap(err, "unable to delete existing terraform.tfvars")
		}
	}

	file, err := os.OpenFile(filename, os.O_TRUNC|os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return errors.Wrap(err, "create tfvars file failed")
	}

	err = p.renderVarsFile(file)
	if err != nil {
		return errors.Wrap(err, "read tfvars file failed")
	}

	err = p.runModule()
	if err != nil {
		return errors.Wrap(err, "run terraform module failed")
	}

	return nil
}

// Delete deletes all resources created by the provider.
func (p *ClusterProvider) Delete() error {
	prov, err := p.Cmd.BinaryExists("terraform")
	if err != nil {
		return errors.Wrap(err, "please install Terraform before running the command")
	}
	if !prov {
		return errors.New("please install Terraform before running the command")
	}

	p.Cmd.SetDirectory(filepath.Join(p.WorkingDir, p.Spec.ModuleSource.Name))
	err = p.Cmd.Run("terraform", "destroy", "-auto-approve", "-input=false")
	if err != nil {
		return errors.Wrap(err, "run terrafrom destroy command failed")
	}

	return nil
}

// GetKubeConfig return kubeconfig file location
func (p *ClusterProvider) GetKubeConfig() (string, error) {
	clusterOutput, err := p.parseTerraformOutput()
	if err != nil {
		return "", errors.Wrap(err, "parsing terraform output failed")
	}

	kubeconfigFile := filepath.Join(p.WorkingDir, p.Spec.ModuleSource.Name, clusterOutput.KubeConfig.Value)
	return kubeconfigFile, nil
}

// returns a new cluster outputs from terraform output.
func (p *ClusterProvider) parseTerraformOutput() (*TerraformOutput, error) {

	// We need to handle this better way, if childern config is missing it will thrown run time exception.
	// To catch invalid memory address or nil pointer dereference, segmentation violation
	if p.Spec.ModuleSource == nil {
		return nil, errors.New("moduleSource need to be initialized for terraform provider, got nil")
	}

	p.Cmd.SetDirectory(filepath.Join(p.WorkingDir, p.Spec.ModuleSource.Name))
	out, err := p.Cmd.Output("terraform", "output", "-json")
	if err != nil {
		return nil, errors.Wrap(err, "command failed").WithField("output", out)
	}

	var terraformOutput TerraformOutput

	err = json.Unmarshal(out, &terraformOutput)

	if err != nil {
		return nil, errors.Wrap(err, "failed to read terraform output")
	}

	return &terraformOutput, nil
}

// New creates a new instance of the ClusterProvider
func New(homeDir string, c *config.ClusterConfig, logger *log.Logger) (*ClusterProvider, error) {

	if c == nil {
		return nil, errors.New("the config object needs to be initialized, got nil")
	}

	var spec InfrastructureSpec

	workingDir, err := providers.CreateProviderDir(homeDir, c.ClusterName, c.Infrastructure.Provider.Name)

	if err != nil {
		return nil, errors.Wrap(err, "failed to create provider working directory in the Yala cluster directory")
	}

	err = viper.UnmarshalKey("infrastructure.provider.spec", &spec)
	if err != nil {
		return nil, errors.Wrap(err, "could not unmarshal infrastructure provider spec to provider.InfrastructureSpec")
	}

	cmdRunner, err := runner.New(logger.Writer(), logger.Writer(), homeDir)

	if err != nil {
		log.WithError(err).WithFields(log.Fields{
			"workingDir": homeDir,
		}).Fatal("Failed to create command runner")
	}

	return &ClusterProvider{
		Name:              c.Infrastructure.Provider.Name,
		ClusterName:       c.ClusterName,
		KubernetesVersion: c.Infrastructure.Provider.KubernetesVersion,
		WorkingDir:        workingDir,
		Spec:              &spec,
		Cmd:               cmdRunner,
	}, nil
}
