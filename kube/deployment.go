package kube

import (
	"context"

	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func deploymentDeploy(ctx context.Context, container *apiv1.Container, client *kubernetes.Clientset, opt *DeployOpt) (err error) {
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:   opt.Name,
			Labels: opt.Labels,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &opt.ReplicaNum,
			Selector: &metav1.LabelSelector{
				MatchLabels: opt.PodLabels,
			},
			Template: apiv1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: opt.PodLabels,
				},
				Spec: apiv1.PodSpec{
					Containers: []apiv1.Container{
						*container,
					},
				},
			},
		},
	}

	// 看看之前部署的deployment还存不存在
	deploymentsClient := client.AppsV1().Deployments(opt.Namespace)
	_, err = deploymentsClient.Get(context.TODO(), opt.Name, metav1.GetOptions{})
	if err != nil {
		if kerrors.IsNotFound(err) {
			// create
			_, err = deploymentsClient.Create(ctx, deployment, metav1.CreateOptions{})
			if err != nil {
				return
			}
		} else {
			return
		}
	} else {
		// 否则，原来的deployment跑的好好的，那么就只需要更新就行
		_, err = deploymentsClient.Update(ctx, deployment, metav1.UpdateOptions{})
		if err != nil {
			return
		}
	}
	return
}
