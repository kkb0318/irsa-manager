package manifests

import (
	"fmt"

	awsclient "github.com/kkb0318/irsa-manager/internal/aws"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

type ServiceAccountBuilder struct {
	annotation map[string]string
}

func NewServiceAccountBuilder() *ServiceAccountBuilder {
	return &ServiceAccountBuilder{}
}

func (b *ServiceAccountBuilder) WithIRSAAnnotation(role awsclient.RoleManager) *ServiceAccountBuilder {
	b.annotation = map[string]string{
		"eks.amazonaws.com/role-arn": fmt.Sprintf("arn:aws:iam::%s:role/%s", role.AccountId, role.RoleName),
	}
	return b
}

func (b *ServiceAccountBuilder) Build(namespacedName types.NamespacedName) *corev1.ServiceAccount {
	return &corev1.ServiceAccount{
		TypeMeta: metav1.TypeMeta{
			APIVersion: corev1.SchemeGroupVersion.String(),
			Kind:       "ServiceAccount",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        namespacedName.Name,
			Namespace:   namespacedName.Namespace,
			Annotations: b.annotation,
		},
	}
}
