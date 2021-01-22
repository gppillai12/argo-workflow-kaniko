package aws

var tfvarsTemplate = `
cluster_name = "{{ .ClusterName }}"
k8s_version  = "{{ .KubernetesVersion }}"
{{ if .Spec.Region }}
region       = "{{ .Spec.Region }}"
{{ end }}
{{ if .Spec.VpcID }}
vpc_id       = "{{ .Spec.VpcID }}"
{{ end }}
working_dir  = "{{ .WorkingDir }}"

{{ if .Spec.PrivateSubnets }}
private_subnets = [
	{{ range .Spec.PrivateSubnets }}
		"{{ . }}",
	{{ end }}
]
{{ end }}

worker_groups = [
	{{ range .Spec.WorkerGroups }}
	{
		name = "{{ .Name }}"
		instance_type = "{{ .InstanceType }}"
		asg_desired_capacity = "{{ .DesiredCapacity }}"
	},
	{{ end }}
]
`
