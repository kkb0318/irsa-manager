package manifests

import (
	"fmt"

	"github.com/kkb0318/irsa-manager/internal/selfhosted"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

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

func (b *SecretBuilder) Build(name, ns string) (*corev1.Secret, error) {
	if b.data == nil {
		return nil, fmt.Errorf("Secret.Data is empty")
	}
	secret := &corev1.Secret{
		ObjectMeta: v1.ObjectMeta{
			Name:      name,
			Namespace: ns,
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
