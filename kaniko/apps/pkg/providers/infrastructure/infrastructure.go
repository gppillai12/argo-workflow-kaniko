package infrastructure

// ClusterProvider is the interface that must be implemented by all Yala infrastructure providers.
type ClusterProvider interface {
	Apply() error
	Delete() error
	GetKubeConfig() (string, error)
}
