package kube

import (
	"context"
	"fmt"
	"strings"
	"time"

	apiv1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	watch2 "k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/watch"
)

const (
	DefaultDuration  = 10 * time.Minute
	DefaultNameSpace = "default"
)

type PodDeployOpt struct {
	KubeConfPath string
	Labels       map[string]string
	extraLabels  map[string]string
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
	labels   map[string]string
	Quota
}

func (opt *PodDeployOpt) fix() {
	if opt.Duration <= 0 {
		opt.Duration = DefaultDuration
	}
	if len(opt.Namespace) <= 0 {
		opt.Namespace = DefaultNameSpace
	}

	// generate pod labels
	podLabels := make(map[string]string)
	for k, v := range opt.Labels {
		podLabels[k] = v
	}
	for k, v := range opt.extraLabels {
		podLabels[k] = v
	}
	opt.spec.labels = podLabels
}

type ErrPodDeploy struct {
	Msg string
}

func (err *ErrPodDeploy) Error() string {
	return err.Msg
}

var ErrPodDeployTimeout = &ErrPodDeploy{Msg: "timeout"}

func PodDeploy(ctx context.Context, opt *PodDeployOpt) (err error) {
	opt.fix()
	ctx, cancel := context.WithTimeout(ctx, opt.Duration)
	defer cancel()

	container, err := getContainer(opt.spec)
	if err != nil {
		return
	}

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

		deployOpt := &DeployOpt{
			Name:       opt.spec.Name,
			Labels:     opt.Labels,
			ReplicaNum: opt.ReplicaNum,
			Namespace:  opt.Namespace,
			PodLabels:  opt.spec.labels,
		}
		var controller PodController

		if opt.Stateful {
			controller = NewStatefulSetController(container, clientSet, deployOpt)
		} else {
			controller = NewDeploymentController(container, clientSet, deployOpt)
		}

		// deploy
		err = controller.DeployOrUpdate(ctx)
		if err != nil {
			return
		}
		// wait pod for done
		//errMsg, err = waitPodsForRunning(ctx, clientSet, opt.Namespace, opt.spec.labels)
		return
	}()

	select {
	case <-ctx.Done():
		err = ErrPodDeployTimeout
	case err = <-errCh:
	}
	return
}

func waitPodForReady(ctx context.Context, client *kubernetes.Clientset, namespace string, labels map[string]string) (errMsg string, err error) {
	listOpt, err := getListOpt(labels)
	if err != nil {
		return err.Error(), err
	}
	condition := func(event watch2.Event) (bool, error) {
		pod, ok := event.Object.(*v1.Pod)
		if !ok {
			return false, nil
		}
		switch pod.Status.Phase {
		case apiv1.PodRunning, apiv1.PodSucceeded:
			return true, nil
		case apiv1.PodFailed:
			errMsg = pod.Status.Message
			return false, nil
		default:
			return false, nil
		}
	}
	client.CoreV1().Pods(namespace).GetLogs(namespace, &v1.PodLogOptions{
		TypeMeta:                     metav1.TypeMeta{},
		Container:                    "",
		Follow:                       false,
		Previous:                     false,
		SinceSeconds:                 nil,
		SinceTime:                    nil,
		Timestamps:                   false,
		TailLines:                    nil,
		LimitBytes:                   nil,
		InsecureSkipTLSVerifyBackend: false,
	})
	lw := &cache.ListWatch{
		ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
			return client.CoreV1().Pods(namespace).List(ctx, *listOpt)
		},
		WatchFunc: func(options metav1.ListOptions) (watch2.Interface, error) {
			return client.CoreV1().Pods(namespace).Watch(ctx, *listOpt)
		},
	}

	watcher, err := watch.NewRetryWatcher("1", lw)
	if err != nil {
		return
	}
	ch := watcher.ResultChan()
	for {
		select {
		case event, ok := <-ch:
			if !ok {
				err = watch.ErrWatchClosed
				return
			}
			if ok, err := condition(event); err != nil {
				return err.Error(), err
			} else if ok {
				return errMsg, nil
			}
		case <-ctx.Done():
			err = ErrPodDeployTimeout
			return
		}
	}
	//_, err = watch.Until(ctx, "1", lw, condition)
	//return
}

func getContainer(spec PodSpec) (*apiv1.Container, error) {
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

	// quota
	resourceList := make(map[v1.ResourceName]resource.Quantity)
	if len(spec.CPU) > 0 {
		quantity, err := resource.ParseQuantity(spec.CPU)
		if err != nil {
			return nil, err
		}
		resourceList[v1.ResourceCPU] = quantity
	}
	if len(spec.Memory) > 0 {
		quantity, err := resource.ParseQuantity(spec.Memory)
		if err != nil {
			return nil, err
		}
		resourceList[v1.ResourceMemory] = quantity
	}
	if len(resourceList) > 0 {
		container.Resources.Limits = resourceList
		container.Resources.Requests = resourceList
	}

	// TODO: 持久化
	return container, nil
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

func listPodsByLabels(ctx context.Context, client *kubernetes.Clientset, labels map[string]string, namespace string) ([]v1.Pod, error) {
	selector, err := metav1.LabelSelectorAsSelector(&metav1.LabelSelector{
		MatchLabels: labels,
	})
	if err != nil {
		return nil, err
	}
	return listPodsBySelector(ctx, client, selector, namespace)
}

func listPodsBySelector(ctx context.Context, client *kubernetes.Clientset, selector labels.Selector, namespace string) ([]v1.Pod, error) {
	podList, err := client.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: selector.String(),
	})
	if err != nil {
		return nil, err
	}
	return podList.Items, nil
}
