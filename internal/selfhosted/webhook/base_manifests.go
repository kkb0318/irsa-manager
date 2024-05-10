package webhook

import (
	"k8s.io/api/admissionregistration/v1beta1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

var mutatingWebhookConfiguration = v1beta1.MutatingWebhookConfiguration{
	TypeMeta: metav1.TypeMeta{
		APIVersion: "admissionregistration.k8s.io/v1beta1",
		Kind:       "MutatingWebhookConfiguration",
	},
	ObjectMeta: metav1.ObjectMeta{
		Name:      "pod-identity-webhook",
		Namespace: "kube-system",
	},
	Webhooks: []v1beta1.MutatingWebhook{
		{
			Name: "pod-identity-webhook.amazonaws.com",
			ClientConfig: v1beta1.WebhookClientConfig{
				Service: &v1beta1.ServiceReference{
					Name:      "pod-identity-webhook",
					Namespace: "kube-system",
					Path:      "/mutate",
				},
				CABundle: []byte("${CA_BUNDLE}"),
			},
			Rules: []v1beta1.RuleWithOperations{
				{
					Operations: []v1beta1.OperationType{"CREATE"},
					Rule: v1beta1.Rule{
						APIGroups:   []string{""},
						APIVersions: []string{"v1"},
						Resources:   []string{"pods"},
					},
				},
			},
			FailurePolicy: (*v1beta1.FailurePolicyType)(nil),
		},
	},
}

var deployment = appsv1.Deployment{
	TypeMeta: metav1.TypeMeta{
		APIVersion: "apps/v1",
		Kind:       "Deployment",
	},
	ObjectMeta: metav1.ObjectMeta{
		Name:      "pod-identity-webhook",
		Namespace: "kube-system",
	},
	Spec: appsv1.DeploymentSpec{
		Replicas: 1,
		Selector: &metav1.LabelSelector{
			MatchLabels: map[string]string{"app": "pod-identity-webhook"},
		},
		Template: corev1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Labels: map[string]string{"app": "pod-identity-webhook"},
			},
			Spec: corev1.PodSpec{
				ServiceAccountName: "pod-identity-webhook",
				Containers: []corev1.Container{
					{
						Name:            "pod-identity-webhook",
						Image:           "quay.io/amis/pod-identity-webhook:v0.0.1",
						ImagePullPolicy: corev1.PullAlways,
						Command:         []string{"/webhook", "--in-cluster", "--namespace=kube-system", "--service-name=pod-identity-webhook", "--tls-secret=pod-identity-webhook", "--annotation-prefix=eks.amazonaws.com", "--token-audience=sts.amazonaws.com", "--logtostderr"},
						VolumeMounts: []corev1.VolumeMount{
							{
								Name:      "webhook-certs",
								MountPath: "/var/run/app/certs",
								ReadOnly:  false,
							},
						},
					},
				},
				Volumes: []corev1.Volume{
					{
						Name: "webhook-certs",
						VolumeSource: corev1.VolumeSource{
							EmptyDir: &corev1.EmptyDirVolumeSource{},
						},
					},
				},
			},
		},
	},
}

var serviceAccount = corev1.ServiceAccount{
	TypeMeta: metav1.TypeMeta{
		APIVersion: "v1",
		Kind:       "ServiceAccount",
	},
	ObjectMeta: metav1.ObjectMeta{
		Name:      "pod-identity-webhook",
		Namespace: "kube-system",
	},
}

type ClusterRole struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Rules             []rbacv1.PolicyRule `json:"rules,omitempty"`
}

var clusterRole = ClusterRole{
	TypeMeta: metav1.TypeMeta{
		APIVersion: "rbac.authorization.k8s.io/v1",
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

type ClusterRoleBinding struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	RoleRef           rbacv1.RoleRef   `json:"roleRef"`
	Subjects          []rbacv1.Subject `json:"subjects"`
}

var clusterRoleBinding = ClusterRoleBinding{
	TypeMeta: metav1.TypeMeta{
		APIVersion: "rbac.authorization.k8s.io/v1",
		Kind:       "ClusterRoleBinding",
	},
	ObjectMeta: metav1.ObjectMeta{
		Name: "pod-identity-webhook",
	},
	RoleRef: rbacv1.RoleRef{
		APIGroup: "rbac.authorization.k8s.io",
		Kind:     "ClusterRole",
		Name:     "pod-identity-webhook",
	},
	Subjects: []rbacv1.Subject{
		{
			Kind:      "ServiceAccount",
			Name:      "pod-identity-webhook",
			Namespace: "kube-system",
		},
	},
}

var service = corev1.Service{
	TypeMeta: metav1.TypeMeta{
		APIVersion: "v1",
		Kind:       "Service",
	},
	ObjectMeta: metav1.ObjectMeta{
		Name:      "pod-identity-webhook",
		Namespace: "kube-system",
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
		Selector: map[string]string{
			"app": "pod-identity-webhook",
		},
	},
}
