package kubernetes

import (
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type KubernetesClient struct {
	client client.Client
	owner  Owner
}
