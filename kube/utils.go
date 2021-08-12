package kube

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func _(i int32) *int32 {
	return &i
}

func getListOpt(labels map[string]string) (*metav1.ListOptions, error) {
	selector, err := metav1.LabelSelectorAsSelector(&metav1.LabelSelector{
		MatchLabels: labels,
	})
	if err != nil {
		return nil, err
	}
	return &metav1.ListOptions{
		LabelSelector: selector.String(),
	}, nil
}
