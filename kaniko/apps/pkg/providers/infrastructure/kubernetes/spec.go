package kubernetes

// ProviderName contains the constant provider name value for 'none'.
const ProviderName string = "kubernetes"

// InfrastructureSpec contains configuration for the provider.
type InfrastructureSpec struct {
	KubeConfig string `json:"kubeConfig"`
}
