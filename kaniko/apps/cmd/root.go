package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"

	homedir "github.com/mitchellh/go-homedir"
	log "github.com/sirupsen/logrus"
)

var cfgFile string

var version = "v0.0.0-dev"

var globalOptions = &GlobalOptions{
	LogFormat:         "text",
	LogLevel:          "info",
	LogTimestamp:      "15:04:05",
	ClusterConfigFile: "./platform.yaml",
}

var rootCmd = &cobra.Command{
	Use:   "yala",
	Short: "Yala manage Mavenir's CI/CD platform",
	Long: `Yala manage Mavenir's CI/CD platform
It simplify deployment process of components like harbor, argocd and helm charts.
`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		initLogger()
		if globalOptions.HomeDir == "" {
			userHomeDir, err := homedir.Dir()
			if err != nil {
				log.WithError(err).Fatal("Failed to find user home directory.")
			}
			globalOptions.HomeDir = fmt.Sprintf("%s/.yala/", userHomeDir)
		} else {
			absHomeDir, err := filepath.Abs(globalOptions.HomeDir)
			if err != nil {
				log.WithError(err).Fatal("Failed to get absolute directory.")
			}
			globalOptions.HomeDir = absHomeDir

			expandedHomeDir, err := homedir.Expand(globalOptions.HomeDir)
			if err != nil {
				log.WithError(err).Fatal("Failed to expand user home directory.")
			}
			globalOptions.HomeDir = expandedHomeDir
		}

		if globalOptions.ClusterConfigFile == "" {
			currentDir, err := os.Getwd()
			if err != nil {
				log.WithError(err).Fatal("Failed to get current directory.")
			}
			globalOptions.ClusterConfigFile = fmt.Sprintf("%s/platform.yaml", currentDir)
		} else {
			absFilepath, err := filepath.Abs(globalOptions.ClusterConfigFile)
			if err != nil {
				log.WithError(err).Fatal("Failed to get absolute filepath.")
			}
			globalOptions.ClusterConfigFile = absFilepath

			expandedFilepath, err := homedir.Expand(globalOptions.ClusterConfigFile)
			if err != nil {
				log.WithError(err).Fatal("Failed to expand user home directory.")
			}
			globalOptions.ClusterConfigFile = expandedFilepath
		}

	},
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

func initLogger() {
	// Setup level

	if level, err := log.ParseLevel(globalOptions.LogLevel); err == nil {
		log.SetLevel(level)
	} else {
		log.WithError(err).WithField("level", globalOptions.LogLevel).Warn("Unknown log level, using default 'info'.")
	}

	if log.GetLevel() > log.InfoLevel {
		log.SetReportCaller(true)
		globalOptions.LogTimestamp = time.RFC3339
	}

	errLogger := log.New()
	errLogger.SetLevel(log.GetLevel())
	errLogger.SetOutput(os.Stderr)
	outLogger := log.New()
	outLogger.SetLevel(log.GetLevel())
	outLogger.SetOutput(os.Stdout)

	// Set formatter
	switch globalOptions.LogFormat {
	case "json":
		log.SetFormatter(&log.JSONFormatter{
			TimestampFormat: time.RFC3339,
		})
		errLogger.SetFormatter(&log.JSONFormatter{
			TimestampFormat: time.RFC3339,
		})
		outLogger.SetFormatter(&log.JSONFormatter{
			TimestampFormat: time.RFC3339,
		})
	case "text":
		log.SetFormatter(&log.TextFormatter{
			TimestampFormat: globalOptions.LogTimestamp,
			FullTimestamp:   true,
		})
		errLogger.SetFormatter(&log.TextFormatter{
			TimestampFormat: globalOptions.LogTimestamp,
			FullTimestamp:   true,
		})
		outLogger.SetFormatter(&log.TextFormatter{
			TimestampFormat: globalOptions.LogTimestamp,
			FullTimestamp:   true,
		})
	default:
		log.SetFormatter(&log.TextFormatter{
			TimestampFormat: globalOptions.LogTimestamp,
			FullTimestamp:   true,
		})
		errLogger.SetFormatter(&log.TextFormatter{
			TimestampFormat: globalOptions.LogTimestamp,
			FullTimestamp:   true,
		})
		outLogger.SetFormatter(&log.TextFormatter{
			TimestampFormat: globalOptions.LogTimestamp,
			FullTimestamp:   true,
		})
		log.WithField("format", globalOptions.LogFormat).Warn("Unknown log format, using default 'text'.")
	}

}

func initGlobalFlags() {
	f := rootCmd.PersistentFlags()

	f.StringVar(&globalOptions.LogLevel, "log-level", globalOptions.LogLevel, "Sets the log level for output. [debug|info|warn|error]")
	f.StringVar(&globalOptions.LogFormat, "log-format", globalOptions.LogFormat, "Sets the log output format. [text|json]")
	f.StringVar(&globalOptions.HomeDir, "home", "", "Sets home directory used for Yala state.")
	f.StringVarP(&globalOptions.ClusterConfigFile, "filename", "f", "", "Sets cluster config file for the desired cluster state.")
}

func init() {
	log.SetOutput(os.Stdout)
	log.SetLevel(log.InfoLevel)
	log.SetFormatter(&log.TextFormatter{
		TimestampFormat: globalOptions.LogTimestamp,
		FullTimestamp:   true,
	})

	initGlobalFlags()

	rootCmd.Version = version
	rootCmd.AddCommand(applyCmd)
	rootCmd.AddCommand(deleteCmd)
	rootCmd.AddCommand(bundleCmd)
	rootCmd.AddCommand(unbundleCmd)

}

// Execute calls the root level Cobra command execute.
// This is the entry point to Yala.
func Execute() error {
	return rootCmd.Execute()
}
