package selfhosted

import (
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type WebHook interface {
	Resources() []client.Object
	Create()
	Delete()
}
