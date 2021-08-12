package kube

import (
	"context"
	"path/filepath"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

// Client is a wrapper for *kubernetes.Clientset
type Client struct {
	*kubernetes.Clientset
	Ctx       context.Context
	NameSpace string
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

func NewClient(ctx context.Context, configPath, namespace string) (*Client, error) {
	clientSet, err := client(configPath)
	if err != nil {
		return nil, err
	}
	return &Client{
		Clientset: clientSet,
		Ctx:       ctx,
		NameSpace: namespace,
	}, nil
}

func MustNewClient(ctx context.Context, configPath, namespace string) *Client {
	return &Client{
		Clientset: mustClient(configPath),
		Ctx:       ctx,
		NameSpace: namespace,
	}
}
