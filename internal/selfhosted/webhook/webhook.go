package webhook

import (
	"fmt"

	"github.com/kkb0318/irsa-manager/internal/manifests"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type WebhookSetup struct {
	resources []client.Object
}

func secretNamespacedName() types.NamespacedName {
	return types.NamespacedName{
		Name:      "pod-identity-webhook",
		Namespace: WEBHOOK_NAMESPACE,
	}
}

func (w *WebhookSetup) Resources() []client.Object {
	return w.resources
}

func NewWebHookSetup() (*WebhookSetup, error) {
	factory := newBaseManifestFactory()
	resources, err := myCertificate(factory)
	if err != nil {
		return nil, err
	}
	return &WebhookSetup{resources}, nil
}

func myCertificate(base *baseManifestFactory) ([]client.Object, error) {
	tlsCredential, err := CreateTlsCredential(serviceNamespacedName())
	if err != nil {
		return nil, err
	}
	resources := []client.Object{}
	secretNamespacedName := secretNamespacedName()
	secret, err := manifests.NewSecretBuilder().
		WithCertificate(tlsCredential).
		Build(secretNamespacedName)
	if err != nil {
		return nil, err
	}

	deploy := base.deployment()
	deploy.Spec.Template.Spec.Containers[0].Command = []string{
		"/webhook",
		"--in-cluster=false",
		fmt.Sprintf("--namespace=%s", WEBHOOK_NAMESPACE),
		fmt.Sprintf("--service-name=%s", base.serviceMeta.Name),
		fmt.Sprintf("--tls-secret=%s", secretNamespacedName.Name),
		"--annotation-prefix=eks.amazonaws.com",
		"--token-audience=sts.amazonaws.com",
		"--logtostderr",
	}
	deploy.Spec.Template.Spec.Volumes = []corev1.Volume{
		{
			Name: "cert",
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName: secretNamespacedName.Name,
				},
			},
		},
	}
	mutate := base.mutatingWebhookConfiguration()
	mutate.Webhooks[0].ClientConfig.CABundle = tlsCredential.Certificate()
	resources = append(resources,
		secret,
		deploy,
		mutate,
		base.clusterRole(),
		base.clusterRoleBinding(),
		base.serviceAccount(),
		base.service(),
	)
	return resources, nil
}
