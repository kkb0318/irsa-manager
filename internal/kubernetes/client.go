package kubernetes

import (
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type KubernetesClient struct {
	client client.Client
	owner  Owner
}

func NewKubernetesClient(c client.Client, owner Owner) (*KubernetesClient, error) {
	return &KubernetesClient{
		client: c,
		owner:  owner,
	}, nil
}
