package kube

import (
	"path/filepath"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

func Int32Ptr(i int32) *int32 {
	return &i
}

func client(confPath string) (*kubernetes.Clientset, error) {
	if confPath == "" {
		confPath = filepath.Join(homedir.HomeDir(), ".kube/config")
	} else {
		confPath, _ = filepath.Abs(confPath)
	}
	config, err := clientcmd.BuildConfigFromFlags("", confPath)
	if err != nil {
		return nil, err
	}
	return kubernetes.NewForConfig(config)
}

func mustClient(confPath string) *kubernetes.Clientset {
	if confPath == "" {
		confPath = filepath.Join(homedir.HomeDir(), ".kube/config")
	} else {
		confPath, _ = filepath.Abs(confPath)
	}
	config, err := clientcmd.BuildConfigFromFlags("", confPath)
	if err != nil {
		panic(err.Error())
	}
	clientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}
	return clientSet
}
