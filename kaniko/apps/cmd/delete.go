package cmd

import (
	"fmt"
	"os"
	"strings"
	"yala/pkg/config"
	"yala/pkg/providers/deployment"
	"yala/pkg/providers/deployment/helm"
	"yala/pkg/providers/infrastructure"
	"yala/pkg/providers/infrastructure/kubernetes"
	"yala/pkg/providers/infrastructure/terraform/aws"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var deleteOptions = &DeleteOptions{}

var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete the cluster specified in the config.",
	Long:  "Delete the cluster specified in the config.",
	Run: func(cmd *cobra.Command, args []string) {
		log.Debug("Running Delete Operation.")
		clusterConfig, err := config.NewFromFile(globalOptions.ClusterConfigFile)
		if err != nil {
			log.WithError(err).WithField("filename", globalOptions.ClusterConfigFile).Fatal("Failed to create cluster config object from config file.")
		}

		logger := log.StandardLogger()

		var provider infrastructure.ClusterProvider

		switch clusterConfig.Infrastructure.Provider.Name {
		case kubernetes.ProviderName:
			provider, err = kubernetes.New(globalOptions.HomeDir, clusterConfig)
		case aws.ProviderName:
			provider, err = aws.New(globalOptions.HomeDir, clusterConfig, logger)
		default:
			log.WithField("provider", clusterConfig.Infrastructure.Provider.Name).Fatal("Unknown infrastructure provider.")
		}

		if err != nil {
			log.WithError(err).WithFields(log.Fields{
				"provider": clusterConfig.Infrastructure.Provider.Name,
				"config":   fmt.Sprintf("%+v", clusterConfig),
			}).Fatalf("Failed to create new provider.")
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

		log.WithField("provider", clusterConfig.Deployment.Provider.Name).Info("Deleting deployments.")

		if err := deploymentProvider.Delete(); err != nil {
			log.WithError(err).WithField("provider", clusterConfig.Deployment.Provider.Name).Fatal("Failed to uninstall deployment.")
		}

		log.WithField("provider", clusterConfig.Infrastructure.Provider.Name).Info("Deleting infrastructure.")
		if err := provider.Delete(); err != nil {
			if strings.Contains(err.Error(), "Provider directory does not exist") {
				log.WithError(err).WithField("provider", clusterConfig.Infrastructure.Provider.Name).Warn("The provider cluster directory does not exist. You may need to manually clean up remaining resources")
			} else {
				log.WithError(err).WithField("provider", clusterConfig.Infrastructure.Provider.Name).Fatal("Failed to delete infrastructure.")
			}
		}

		log.WithField("directory", globalOptions.ClusterDir).Info("Deleting cluster state.")
		if err := os.RemoveAll(globalOptions.ClusterDir); err != nil {
			log.WithError(err).WithField("directory", globalOptions.ClusterDir).Fatal("Failed to delete provider directory.")
		}

		log.Info("Yala has successfully deleted the resources.")
	},
}

func init() {
}
