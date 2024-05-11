package webhook

import (
	"os"
	"testing"

	"github.com/goccy/go-yaml"
	"github.com/stretchr/testify/assert"
	regv1 "k8s.io/api/admissionregistration/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func TestBaseManifests(t *testing.T) {
	b := newBaseManifestFactory()
	tests := []struct {
		name         string
		runFunc      func() client.Object
		expected     string
		expectedFunc func() client.Object
	}{
		{
			name: "mutatingwebhookconfiguration",
			runFunc: func() client.Object {
				return b.mutatingWebhookConfiguration()
			},
			expected:     "testdata/mutatingwebhook.yaml",
			expectedFunc: testMutatingWebhookConfiguration,
		},
		{
			name: "service",
			runFunc: func() client.Object {
				return b.service()
			},
			expected:     "testdata/service.yaml",
			expectedFunc: testService,
		},
		{
			name: "deployment",
			runFunc: func() client.Object {
				return b.deployment()
			},
			expected:     "testdata/deployment.yaml",
			expectedFunc: testDeployment,
		},
		{
			name: "serviceaccount",
			runFunc: func() client.Object {
				return b.serviceAccount()
			},
			expected:     "testdata/serviceaccount.yaml",
			expectedFunc: testServiceAccount,
		},
		{
			name: "clusterrole",
			runFunc: func() client.Object {
				return b.clusterRole()
			},
			expected:     "testdata/clusterrole.yaml",
			expectedFunc: testClusterRole,
		},
		{
			name: "clusterrolebinding",
			runFunc: func() client.Object {
				return b.clusterRoleBinding()
			},
			expected:     "testdata/clusterrolebinding.yaml",
			expectedFunc: testClusterRoleBinding,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := tt.runFunc()
			data, err := os.ReadFile(tt.expected)
			assert.NoError(t, err)
			expected := tt.expectedFunc()
			err = yaml.UnmarshalWithOptions(data, expected, yaml.UseJSONUnmarshaler())
			assert.NoError(t, err)
			assert.Equal(t, expected, actual)
		})
	}
}

func testMutatingWebhookConfiguration() client.Object {
	return &regv1.MutatingWebhookConfiguration{}
}

func testService() client.Object {
	return &corev1.Service{}
}

func testDeployment() client.Object {
	return &appsv1.Deployment{}
}

func testServiceAccount() client.Object {
	return &corev1.ServiceAccount{}
}

func testClusterRole() client.Object {
	return &rbacv1.ClusterRole{}
}

func testClusterRoleBinding() client.Object {
	return &rbacv1.ClusterRoleBinding{}
}
