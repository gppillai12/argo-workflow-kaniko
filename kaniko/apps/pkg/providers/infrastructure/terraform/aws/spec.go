package aws

const ProviderName string = "terraform-aws"

type InfrastructureSpec struct {
	Region         string                 `json:"region"`
	ModuleSource   *TerraformModuleSource `json:"moduleSource"`
	VpcID          string                 `json:"vpcID"`
	PrivateSubnets []string               `json:"privateSubnets"`
	WorkerGroups   *[]WorkerGroup         `json:"workerGroups"`
}

type TerraformModuleSource struct {
	Name string `json:"name"`
	Path string `json:"path"`
}

type WorkerGroup struct {
	Name            string `json:"name"`
	InstanceType    string `json:"instanceType"`
	DesiredCapacity int    `json:"desiredCapacity"`
}

// TerraformOutput is the definition of all Output  provided by terraform out for cluster components.
type TerraformOutput struct {
	// kubeconfig file
	KubeConfig KubeConfig `json:"kubeconfig"`
}

type KubeConfig struct {
	Value string `json:"value"`
}
