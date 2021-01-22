package helm

import (
	errors "github.com/icelolly/go-errors"

	"fmt"
)

const ProviderName string = "helm"

// DeploymentSpec contains configuration for the provider.
type DeploymentSpec struct {
	Charts []map[string]interface{} `json:"charts"`
}

// NewDeploymentSpec creates a DeploymentSpec from an interface with the correct fields.
func NewDeploymentSpec(data interface{}) (*DeploymentSpec, error) {
	// See if type is already compatible
	if deploymentSpec, ok := data.(DeploymentSpec); ok {
		return &deploymentSpec, nil
	}
	if deploymentSpec, ok := data.(*DeploymentSpec); ok {
		return deploymentSpec, nil
	}

	// See if type can be converted to DeploymentSpec
	datamap, ok := data.(map[string]interface{})
	if !ok {
		return nil, errors.New("casting error").WithFields("data", data, "target type", "helm.DeploymentSpec", "source type", fmt.Sprintf("%T", data))
	}

	if _, ok := datamap["charts"]; !ok {
		return nil, errors.New("casting error: cannot cast to helm.DeploymentSpec, missing field 'charts'").WithField("data", data)
	}

	maps, ok := datamap["charts"].([]interface{})
	if !ok {
		return nil, errors.New("casting error: field 'charts' is not a list").WithField("charts", datamap["charts"])
	}

	charts := make([]map[string]interface{}, 0)

	for _, e := range maps {
		newmap := make(map[string]interface{})
		mapymap := e.(map[interface{}]interface{})
		for k, v := range mapymap {
			newmap[k.(string)] = v
		}
		charts = append(charts, newmap)
	}

	return &DeploymentSpec{
		Charts: charts,
	}, nil
}
