package kube

type DeployOpt struct {
	Name       string
	Labels     map[string]string
	ReplicaNum int32
	Namespace  string
	PodLabels  map[string]string
}

type Port struct {
	Name     string
	Protocol string
	Port     int32
}

type Cmd struct {
	Command []string
	Args    []string
}
