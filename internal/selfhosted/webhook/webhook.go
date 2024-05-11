package webhook

import (
	"fmt"

	"github.com/kkb0318/irsa-manager/internal/manifests"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type AwsWebhook struct {
	resources []client.Object
}

func secretNamespacedName() types.NamespacedName {
	return types.NamespacedName{
		Name:      "pod-identity-webhook",
		Namespace: WEBHOOK_NAMESPACE,
	}
}

func NewWebHook() (*AwsWebhook, error) {
	factory := newBaseManifestFactory()
	resources, err := myCertificate(factory)
	if err != nil {
		return nil, err
	}
	return &AwsWebhook{resources}, nil
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
		"--in-cluster",
		fmt.Sprintf("--namespace=%s", WEBHOOK_NAMESPACE),
		fmt.Sprintf("--service-name=%s", base.serviceMeta.Name),
		fmt.Sprintf("--tls-secret=%s", secretNamespacedName.Name),
		"--annotation-prefix=eks.amazonaws.com",
		"--token-audience=sts.amazonaws.com",
		"--logtostderr",
	}
	mutate := base.mutatingWebhookConfiguration()
	mutate.Webhooks[0].ClientConfig.CABundle = []byte(tlsCredential.CaBundle())
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
