package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"yala/pkg/config"
	"yala/pkg/providers/deployment"
	"yala/pkg/providers/deployment/helm"
	"yala/pkg/providers/infrastructure"
	"yala/pkg/providers/infrastructure/kubernetes"
	"yala/pkg/providers/infrastructure/terraform/aws"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var applyOptions = &ApplyOptions{}

var applyCmd = &cobra.Command{
	Use:   "apply",
	Short: "Apply the cluster configuration to create, update, or repair a insfrastructure.",
	Long:  "Apply the cluster configuration to create, update, or repair a insfrastructure.",
	Run: func(cmd *cobra.Command, args []string) {
		log.Debug("Running Apply Operation.")
		clusterConfig, err := config.NewFromFile(globalOptions.ClusterConfigFile)
		if err != nil {
			log.WithError(err).WithField("filename", globalOptions.ClusterConfigFile).Fatal("Failed to create cluster config object from config file.")
		}
		initClusterDir(clusterConfig.ClusterName)
		logger := log.StandardLogger()

		var provider infrastructure.ClusterProvider
		switch clusterConfig.Infrastructure.Provider.Name {
		case aws.ProviderName:
			provider, err = aws.New(globalOptions.HomeDir, clusterConfig, logger)
		case kubernetes.ProviderName:
			provider, err = kubernetes.New(globalOptions.HomeDir, clusterConfig)
		default:
			log.WithField("provider", clusterConfig.Infrastructure.Provider.Name).Fatal("Unknown infrastructure provider.")
		}

		if err != nil {
			log.WithError(err).WithFields(log.Fields{
				"provider": clusterConfig.Infrastructure.Provider.Name,
				"config":   fmt.Sprintf("%+v", clusterConfig),
			}).Fatalf("Failed to create new provider.")
		}

		log.WithField("provider", clusterConfig.Infrastructure.Provider.Name).Info("Creating infrastructure.")

		if err := provider.Apply(); err != nil {
			log.WithError(err).WithField("provider", clusterConfig.Infrastructure.Provider.Name).Fatal("Failed to provision infrastructure.")
		}

		kubeConfig, err := provider.GetKubeConfig()

		if err != nil {
			log.WithError(err).WithField("provider", clusterConfig.Infrastructure.Provider.Name).Fatal("Unable to fetch kubeconfig file.")
		}

		var deploymentProvider deployment.ClusterProvider
		switch clusterConfig.Deployment.Provider.Name {
		case helm.ProviderName:
			deploymentProvider, err = helm.New(globalOptions.HomeDir, kubeConfig, clusterConfig, logger)
		default:
			log.WithField("provider", clusterConfig.Infrastructure.Provider.Name).Fatal("Unknown deployment provider.")
		}

		if err != nil {
			log.WithError(err).WithFields(log.Fields{
				"provider": clusterConfig.Deployment.Provider.Name,
				"config":   fmt.Sprintf("%+v", clusterConfig),
			}).Fatalf("Failed to create new provider.")
		}

		log.WithField("provider", clusterConfig.Deployment.Provider.Name).Info("Creating infrastructure.")

		if err := deploymentProvider.Apply(); err != nil {
			log.WithError(err).WithField("provider", clusterConfig.Deployment.Provider.Name).Fatal("Failed to provision infrastructure.")
		}

		log.Info("Yala has successfully applied the cluster configuration.")
	},
}

func initClusterDir(clusterName string) {
	globalOptions.ClusterDir = filepath.Join(globalOptions.HomeDir, clusterName)
	log.WithField("directory", globalOptions.ClusterDir).Info("Creating Yala cluster directory.")

	// MkdirAll will only create dirs if they don't exist and won't error if they do.
	if err := os.MkdirAll(globalOptions.ClusterDir, 0750); err != nil {
		log.WithError(err).WithField("directory", globalOptions.ClusterDir).Fatal("Failed to create Yala cluster directory.")
	}

	backupFilename := filepath.Join(globalOptions.ClusterDir, "platform.yaml")
	log.WithFields(log.Fields{
		"source":      globalOptions.ClusterConfigFile,
		"destination": backupFilename,
	}).Debug("Creating Yala cluster config file backup.")

	// Make a copy of the cluster configuration file. This could be used to diff before an apply.
	f1, err := os.Open(globalOptions.ClusterConfigFile)
	if err != nil {
		log.WithError(err).WithField("filename", globalOptions.ClusterConfigFile).Fatal("Failed to open new cluster config file.")
	}
	defer f1.Close()

	f2, err := os.Create(backupFilename)
	if err != nil {
		log.WithError(err).WithField("filename", globalOptions.ClusterConfigFile).Fatal("Failed to open backup cluster config file.")
	}
	defer f2.Close()

	if _, err = io.Copy(f2, f1); err != nil {
		log.WithError(err).WithFields(log.Fields{
			"source":      globalOptions.ClusterConfigFile,
			"destination": backupFilename,
		}).Fatal("Failed to backup cluster config file.")
	}
}
