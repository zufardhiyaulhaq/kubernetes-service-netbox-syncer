package client

import (
	"log"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type KubernetesClient struct {
	k8sClient *kubernetes.Clientset
}

func (c *KubernetesClient) Client() *kubernetes.Clientset {
	return c.k8sClient
}

func NewKubernetesClient() (*KubernetesClient, error) {
	inClusterConf, err := rest.InClusterConfig()
	if err != nil {
		log.Print("failed when get in-cluster config")
		return nil, err
	}
	k8sClient, err := kubernetes.NewForConfig(inClusterConf)
	if err != nil {
		return nil, err
	}
	conf := KubernetesClient{
		k8sClient: k8sClient,
	}
	return &conf, nil
}
