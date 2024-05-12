package webhook

import (
	regv1 "k8s.io/api/admissionregistration/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
)

type baseManifestFactory struct {
	deploymentMeta                   types.NamespacedName
	serviceMeta                      types.NamespacedName
	serviceAccountMeta               types.NamespacedName
	mutatingWebhookConfigurationMeta types.NamespacedName
	podLabel                         map[string]string
}

func serviceNamespacedName() types.NamespacedName {
	return types.NamespacedName{
		Name:      "pod-identity-webhook",
		Namespace: WEBHOOK_NAMESPACE,
	}
}

const WEBHOOK_NAMESPACE = "kube-system"

func newBaseManifestFactory() *baseManifestFactory {
	return &baseManifestFactory{
		deploymentMeta: types.NamespacedName{
			Name:      "pod-identity-webhook",
			Namespace: WEBHOOK_NAMESPACE,
		},
		serviceMeta: serviceNamespacedName(),
		serviceAccountMeta: types.NamespacedName{
			Name:      "pod-identity-webhook",
			Namespace: WEBHOOK_NAMESPACE,
		},
		mutatingWebhookConfigurationMeta: types.NamespacedName{
			Name:      "pod-identity-webhook",
			Namespace: WEBHOOK_NAMESPACE,
		},
		podLabel: map[string]string{"app": "pod-identity-webhook"},
	}
}

func (b *baseManifestFactory) mutatingWebhookConfiguration() *regv1.MutatingWebhookConfiguration {
	path := "/mutate"
	failurePolicy := regv1.Ignore
	sideEffects := regv1.SideEffectClassNone
	return &regv1.MutatingWebhookConfiguration{
		TypeMeta: metav1.TypeMeta{
			APIVersion: regv1.SchemeGroupVersion.String(),
			Kind:       "MutatingWebhookConfiguration",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      b.mutatingWebhookConfigurationMeta.Name,
			Namespace: b.mutatingWebhookConfigurationMeta.Namespace,
		},
		Webhooks: []regv1.MutatingWebhook{
			{
				Name: "pod-identity-webhook.amazonaws.com",
				ClientConfig: regv1.WebhookClientConfig{
					Service: &regv1.ServiceReference{
						Name:      b.serviceMeta.Name,
						Namespace: b.serviceMeta.Namespace,
						Path:      &path,
					},
				},
				Rules: []regv1.RuleWithOperations{
					{
						Operations: []regv1.OperationType{"CREATE"},
						Rule: regv1.Rule{
							APIGroups:   []string{""},
							APIVersions: []string{"v1"},
							Resources:   []string{"pods"},
						},
					},
				},
				FailurePolicy:           &failurePolicy,
				SideEffects:             &sideEffects,
				AdmissionReviewVersions: []string{"v1beta1"},
			},
		},
	}
}

func (b *baseManifestFactory) deployment() *appsv1.Deployment {
	replicas := int32(1)
	return &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			APIVersion: appsv1.SchemeGroupVersion.String(),
			Kind:       "Deployment",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      b.deploymentMeta.Name,
			Namespace: b.deploymentMeta.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: b.podLabel,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: b.podLabel,
				},
				Spec: corev1.PodSpec{
					ServiceAccountName: b.serviceAccountMeta.Name,
					Containers: []corev1.Container{
						{
							Name:            "pod-identity-webhook",
							Image:           "quay.io/amis/pod-identity-webhook:v0.0.1",
							ImagePullPolicy: corev1.PullAlways,
							// Command:         []string{}, // Command must be patched
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "cert",
									MountPath: "/etc/webhook/certs",
									ReadOnly:  true,
								},
							},
						},
					},
					// Volumes: []corev1.Volume{ //Volumes must be patched
					// 	{
					// 		Name: "cert",
					// 	},
					// },
				},
			},
		},
	}
}

func (b *baseManifestFactory) serviceAccount() *corev1.ServiceAccount {
	return &corev1.ServiceAccount{
		TypeMeta: metav1.TypeMeta{
			APIVersion: corev1.SchemeGroupVersion.String(),
			Kind:       "ServiceAccount",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      b.serviceAccountMeta.Name,
			Namespace: b.serviceAccountMeta.Namespace,
		},
	}
}

func (b *baseManifestFactory) clusterRole() *rbacv1.ClusterRole {
	return &rbacv1.ClusterRole{
		TypeMeta: metav1.TypeMeta{
			APIVersion: rbacv1.SchemeGroupVersion.String(),
			Kind:       "ClusterRole",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "pod-identity-webhook",
		},
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups: []string{""},
				Resources: []string{"secrets"},
				Verbs:     []string{"create", "get", "update", "patch"},
			},
			{
				APIGroups: []string{""},
				Resources: []string{"serviceaccounts"},
				Verbs:     []string{"get", "watch", "list"},
			},
			{
				APIGroups: []string{"certificates.k8s.io"},
				Resources: []string{"certificatesigningrequests"},
				Verbs:     []string{"create", "get", "list", "watch"},
			},
		},
	}
}

func (b *baseManifestFactory) clusterRoleBinding() *rbacv1.ClusterRoleBinding {
	return &rbacv1.ClusterRoleBinding{
		TypeMeta: metav1.TypeMeta{
			APIVersion: rbacv1.SchemeGroupVersion.String(),
			Kind:       "ClusterRoleBinding",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "pod-identity-webhook",
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: rbacv1.SchemeGroupVersion.Group,
			Kind:     "ClusterRole",
			Name:     "pod-identity-webhook",
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      b.serviceAccountMeta.Name,
				Namespace: b.serviceAccountMeta.Namespace,
			},
		},
	}
}

func (b *baseManifestFactory) service() *corev1.Service {
	return &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			APIVersion: corev1.SchemeGroupVersion.String(),
			Kind:       "Service",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      b.serviceMeta.Name,
			Namespace: b.serviceMeta.Namespace,
			Annotations: map[string]string{
				"prometheus.io/port":   "443",
				"prometheus.io/scheme": "https",
				"prometheus.io/scrape": "true",
			},
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Port:       443,
					TargetPort: intstr.FromInt(443),
				},
			},
			Selector: b.podLabel,
		},
	}
}
