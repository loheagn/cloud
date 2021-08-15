package kube

import (
	"context"

	v1 "k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

type ServiceOpt struct {
	BaseServiceOpt
	Type v1.ServiceType
}

type BaseServiceOpt struct {
	Name string
	// Labels 表示选取哪些Pod
	Labels map[string]string
	Ports  []Port
}

func (cli *Client) CreateOrReplaceService(ctx context.Context, opt *ServiceOpt) (svc *v1.Service, err error) {
	service := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:   opt.Name,
			Labels: opt.Labels,
		},

		Spec: v1.ServiceSpec{
			Type:     opt.Type,
			Selector: opt.Labels,
			Ports:    getSvePorts(opt.Ports),
		},
	}

	serviceClient := cli.CoreV1().Services(cli.namespace)
	// 先检查是不是存在之前的service
	_, err = serviceClient.Get(ctx, opt.Name, metav1.GetOptions{})
	if err != nil && !kerrors.IsNotFound(err) {
		return
	}
	if err == nil {
		// 删除原来的service
		deletePolicy := metav1.DeletePropagationForeground
		if err = serviceClient.Delete(ctx, opt.Name, metav1.DeleteOptions{
			PropagationPolicy: &deletePolicy,
		}); err != nil {
			return
		}
	}

	// create
	_, err = serviceClient.Create(ctx, service, metav1.CreateOptions{})
	if err != nil {
		return
	}
	return serviceClient.Get(ctx, opt.Name, metav1.GetOptions{})
}

func getSvePorts(ports []Port) []v1.ServicePort {
	containerPorts := getContainerPorts(ports)
	var svcPorts []v1.ServicePort
	for _, port := range containerPorts {
		p := v1.ServicePort{
			Name:     port.Name,
			Protocol: port.Protocol,
			Port:     port.ContainerPort,
			TargetPort: intstr.IntOrString{
				Type:   intstr.Int,
				IntVal: port.ContainerPort,
			},
		}
		svcPorts = append(svcPorts, p)
	}
	return svcPorts
}
