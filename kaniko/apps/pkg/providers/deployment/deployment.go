package deployment

// ClusterProvider is the interface that must be implemented by all Yala infrastructure providers.
type ClusterProvider interface {
	Apply() error
	Delete() error
}
