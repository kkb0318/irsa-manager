package selfhosted

import (
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Webhook interface {
	Resources() []client.Object
}
