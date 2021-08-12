package kube

import (
	"context"
	"time"

	v1 "k8s.io/api/core/v1"
)

const (
	DefaultDuration  = 10 * time.Minute
	DefaultNameSpace = "default"
)

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

type PodController interface {
	DeployOrUpdate(ctx context.Context) error
	GetPods(ctx context.Context) ([]v1.Pod, error)
	Delete(ctx context.Context) error
	Exists(ctx context.Context) (bool, error)
}
