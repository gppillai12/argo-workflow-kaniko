package cmd

type ApplyOptions struct {
}

type DeleteOptions struct {
}

type GlobalOptions struct {
	LogFormat         string
	LogLevel          string
	LogTimestamp      string
	HomeDir           string
	ClusterDir        string
	ClusterConfigFile string
	BundleDir         string
}
