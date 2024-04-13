package manifests

import (
	"github.com/kkb0318/irsa-manager/internal/selfhosted"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

func SshKeyNamespacedName() types.NamespacedName {
	return types.NamespacedName{
		Namespace: "kube-system",
		Name:      "irsa-manager-key",
	}
}

type SecretBuilder struct {
	data       map[string][]byte
	secretType corev1.SecretType
}

func NewSecretBuilder() *SecretBuilder {
	return &SecretBuilder{
		secretType: corev1.SecretTypeOpaque,
	}
}

func (b *SecretBuilder) WithSSHKey(keyPair selfhosted.KeyPair) *SecretBuilder {
	b.data = map[string][]byte{
		"ssh-publickey":          keyPair.PublicKey(),
		corev1.SSHAuthPrivateKey: keyPair.PrivateKey(),
	}
	b.secretType = corev1.SecretTypeSSHAuth
	return b
}

func (b *SecretBuilder) Build(namespacedName types.NamespacedName) (*corev1.Secret, error) {
	secret := &corev1.Secret{
		ObjectMeta: v1.ObjectMeta{
			Name:      namespacedName.Name,
			Namespace: namespacedName.Namespace,
		},
		TypeMeta: v1.TypeMeta{
			APIVersion: corev1.SchemeGroupVersion.String(),
			Kind:       "Secret",
		},
		Type: b.secretType,
		Data: b.data,
	}
	return secret, nil
}
