package kube

import (
	"context"
	"fmt"
	"strings"
	"time"

	apiv1 "k8s.io/api/core/v1"
)

const (
	DefaultDuration  = 10 * time.Minute
	DefaultNameSpace = "default"
)

type PodDeployOpt struct {
	KubeConfPath string
	Labels       map[string]string
	ReplicaNum   int32
	Stateful     bool
	Namespace    string
	Duration     time.Duration
	spec         PodSpec
}

type PodSpec struct {
	Name     string
	ImageTag string
	Envs     map[string]string
	Ports    []Port
	WorkDir  string
	Cmd      Cmd
	Labels   map[string]string
}

func (opt *PodDeployOpt) fix() {
	if opt.Duration <= 0 {
		opt.Duration = DefaultDuration
	}
	if len(opt.Namespace) <= 0 {
		opt.Namespace = DefaultNameSpace
	}
}

func PodDeploy(ctx context.Context, opt *PodDeployOpt) (err error) {
	opt.fix()
	ctx, cancel := context.WithTimeout(ctx, opt.Duration)
	defer cancel()

	container := getContainer(opt.spec)

	errCh := make(chan error)

	go func() {
		var err error
		defer func() {
			errCh <- err
		}()
		clientSet, err := client(opt.KubeConfPath)
		if err != nil {
			return
		}
		switch opt.Stateful {
		case false:
			err = deploymentDeploy(ctx, container, clientSet, &DeployOpt{
				Name:       opt.spec.Name,
				Labels:     opt.Labels,
				ReplicaNum: opt.ReplicaNum,
			})
		}
	}()

	select {
	case <-ctx.Done():
		err = ctx.Err()
	case err = <-errCh:
	}
	return
}

func getContainer(spec PodSpec) *apiv1.Container {
	container := &apiv1.Container{
		Name:  spec.Name,
		Image: spec.ImageTag,
	}

	// 端口
	container.Ports = getContainerPorts(spec.Ports)

	// 环境变量
	var envs []apiv1.EnvVar
	for k, v := range spec.Envs {
		envs = append(envs, apiv1.EnvVar{
			Name:  k,
			Value: v,
		})
	}
	container.Env = envs

	// workingDir
	if len(spec.WorkDir) > 0 {
		container.WorkingDir = spec.WorkDir
	}

	// command
	if len(spec.Cmd.Command) > 0 {
		container.Command = spec.Cmd.Command
	}
	if len(spec.Cmd.Args) > 0 {
		container.Args = spec.Cmd.Args
	}

	// TODO: 持久化
	return container
}

func getContainerPorts(dbPorts []Port) []apiv1.ContainerPort {
	var ports []apiv1.ContainerPort
	for i, port := range dbPorts {
		p := apiv1.ContainerPort{}

		// 处理端口名称
		if port.Name != "" {
			p.Name = port.Name
		} else {
			p.Name = fmt.Sprintf("port-%d", i)
		}

		// 处理端口协议，默认为tcp
		var protocol apiv1.Protocol
		switch strings.ToLower(port.Protocol) {
		case "udp":
			protocol = apiv1.ProtocolUDP
		case "sctp":
			protocol = apiv1.ProtocolSCTP
		default:
			protocol = apiv1.ProtocolTCP
		}
		p.Protocol = protocol

		//  处理端口号
		p.ContainerPort = port.Port

		ports = append(ports, p)
	}
	return ports
}

//func waitPodForRunning(ctx context.Context)
