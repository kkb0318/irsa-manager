/*
Copyright 2024.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	awsclient "github.com/kkb0318/irsa-manager/internal/aws"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	irsav1alpha1 "github.com/kkb0318/irsa-manager/api/v1alpha1"
)

var _ = Describe("IRSASetup Controller", func() {
	Context("When reconciling a resource", func() {
		tests := []struct {
			name string
			obj  *irsav1alpha1.IRSASetup
			f    func(*IRSASetupReconciler, *irsav1alpha1.IRSASetup)
		}{
			{
				name: "should reconcile successfully",
				obj: &irsav1alpha1.IRSASetup{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-resource1",
						Namespace: "default",
					},
					Spec: irsav1alpha1.IRSASetupSpec{
						Cleanup: true,
						// Mode:    "selfhosted",
						Discovery: irsav1alpha1.Discovery{
							S3: irsav1alpha1.S3Discovery{
								Region:     "ap-northeast-1",
								BucketName: "irsa-manager-1",
							},
						},
					},
				},
				f: func(r *IRSASetupReconciler, obj *irsav1alpha1.IRSASetup) {
					expected := []expectedResource{
						{
							NamespacedName: types.NamespacedName{Name: "irsa-manager-key", Namespace: "kube-system"},
							f:              newSecret,
						},
						// webhook
						{
							NamespacedName: types.NamespacedName{Name: "pod-identity-webhook", Namespace: "kube-system"},
							f:              newDeployment,
						},
						{
							NamespacedName: types.NamespacedName{Name: "pod-identity-webhook", Namespace: "kube-system"},
							f:              newService,
						},
						{
							NamespacedName: types.NamespacedName{Name: "pod-identity-webhook", Namespace: "kube-system"},
							f:              newMutatingWebhookConfiguration,
						},
						{
							NamespacedName: types.NamespacedName{Name: "pod-identity-webhook", Namespace: "kube-system"},
							f:              newServiceAccount,
						},
						{
							NamespacedName: types.NamespacedName{Name: "pod-identity-webhook", Namespace: "kube-system"},
							f:              newClusterRole,
						},
						{
							NamespacedName: types.NamespacedName{Name: "pod-identity-webhook", Namespace: "kube-system"},
							f:              newClusterRoleBinding,
						},
					}

					By("Reconciling the created resource")
					typeNamespacedName := types.NamespacedName{
						Name:      obj.Name,
						Namespace: obj.Namespace,
					}

					_, err := r.Reconcile(ctx, reconcile.Request{
						NamespacedName: typeNamespacedName,
					})
					Expect(err).NotTo(HaveOccurred())
					_, err = r.Reconcile(ctx, reconcile.Request{
						NamespacedName: typeNamespacedName,
					})
					Expect(err).NotTo(HaveOccurred())
					for _, expect := range expected {
						checkExist(expect)
					}
					By("removing the custom resource for the Kind")
					Eventually(func() error {
						return k8sClient.Delete(ctx, obj)
					}, timeout).Should(Succeed())
					_, err = r.Reconcile(ctx, reconcile.Request{
						NamespacedName: typeNamespacedName,
					})
					Expect(err).To(Not(HaveOccurred()))
					for _, expect := range expected {
						checkNoExist(expect)
					}
				},
			},
			{
				name: "error case",
				obj: &irsav1alpha1.IRSASetup{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-resource2",
						Namespace: "default",
					},
					Spec: irsav1alpha1.IRSASetupSpec{
						Cleanup: true,
						// Mode:    "selfhoted",
						Discovery: irsav1alpha1.Discovery{
							S3: irsav1alpha1.S3Discovery{
								Region:     "ap-northeast-1",
								BucketName: "irsa-manager-1",
							},
						},
					},
				},
				f: func(r *IRSASetupReconciler, obj *irsav1alpha1.IRSASetup) {
					expected := []expectedResource{
						{
							NamespacedName: types.NamespacedName{Name: "irsa-manager-key", Namespace: "kube-system"},
							f:              newSecret,
						},
						// webhook
						{
							NamespacedName: types.NamespacedName{Name: "pod-identity-webhook", Namespace: "kube-system"},
							f:              newDeployment,
						},
						{
							NamespacedName: types.NamespacedName{Name: "pod-identity-webhook", Namespace: "kube-system"},
							f:              newService,
						},
						{
							NamespacedName: types.NamespacedName{Name: "pod-identity-webhook", Namespace: "kube-system"},
							f:              newMutatingWebhookConfiguration,
						},
						{
							NamespacedName: types.NamespacedName{Name: "pod-identity-webhook", Namespace: "kube-system"},
							f:              newServiceAccount,
						},
						{
							NamespacedName: types.NamespacedName{Name: "pod-identity-webhook", Namespace: "kube-system"},
							f:              newClusterRole,
						},
						{
							NamespacedName: types.NamespacedName{Name: "pod-identity-webhook", Namespace: "kube-system"},
							f:              newClusterRoleBinding,
						},
					}
					typeNamespacedName := types.NamespacedName{
						Name:      obj.Name,
						Namespace: obj.Namespace,
					}
					By("secret does not exist when reconciling with the AwsClient error")
					r.AwsClient = newMockAwsClient(&mockAwsIamAPI{createOidcErr: fmt.Errorf("createOidcErr")}, &mockAwsS3API{}, &mockAwsStsAPI{})
					_, err := r.Reconcile(ctx, reconcile.Request{
						NamespacedName: typeNamespacedName,
					})
					Expect(err).To(HaveOccurred())
					for _, expect := range expected {
						checkNoExist(expect)
					}
					By("successfully Reconciling")
					r.AwsClient = newMockAwsClient(&mockAwsIamAPI{}, &mockAwsS3API{}, &mockAwsStsAPI{})
					_, err = r.Reconcile(ctx, reconcile.Request{
						NamespacedName: typeNamespacedName,
					})
					Expect(err).NotTo(HaveOccurred())
					for _, expect := range expected {
						checkExist(expect)
					}
					By("removing the custom resource for the Kind")
					Eventually(func() error {
						return k8sClient.Delete(ctx, obj)
					}, timeout).Should(Succeed())
					_, err = r.Reconcile(ctx, reconcile.Request{
						NamespacedName: typeNamespacedName,
					})
					Expect(err).To(Not(HaveOccurred()))
					for _, expect := range expected {
						checkNoExist(expect)
					}
				},
			},
			{
				name: "no cleanup",
				obj: &irsav1alpha1.IRSASetup{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-resource2",
						Namespace: "default",
					},
					Spec: irsav1alpha1.IRSASetupSpec{
						Cleanup: false,
						// Mode:    "selfhoted",
						Discovery: irsav1alpha1.Discovery{
							S3: irsav1alpha1.S3Discovery{
								Region:     "ap-northeast-1",
								BucketName: "irsa-manager-1",
							},
						},
					},
				},
				f: func(r *IRSASetupReconciler, obj *irsav1alpha1.IRSASetup) {
					expected := []expectedResource{
						{
							NamespacedName: types.NamespacedName{Name: "irsa-manager-key", Namespace: "kube-system"},
							f:              newSecret,
						},
						// webhook
						{
							NamespacedName: types.NamespacedName{Name: "pod-identity-webhook", Namespace: "kube-system"},
							f:              newDeployment,
						},
						{
							NamespacedName: types.NamespacedName{Name: "pod-identity-webhook", Namespace: "kube-system"},
							f:              newService,
						},
						{
							NamespacedName: types.NamespacedName{Name: "pod-identity-webhook", Namespace: "kube-system"},
							f:              newMutatingWebhookConfiguration,
						},
						{
							NamespacedName: types.NamespacedName{Name: "pod-identity-webhook", Namespace: "kube-system"},
							f:              newServiceAccount,
						},
						{
							NamespacedName: types.NamespacedName{Name: "pod-identity-webhook", Namespace: "kube-system"},
							f:              newClusterRole,
						},
						{
							NamespacedName: types.NamespacedName{Name: "pod-identity-webhook", Namespace: "kube-system"},
							f:              newClusterRoleBinding,
						},
					}
					typeNamespacedName := types.NamespacedName{
						Name:      obj.Name,
						Namespace: obj.Namespace,
					}
					_, err := r.Reconcile(ctx, reconcile.Request{
						NamespacedName: typeNamespacedName,
					})
					Expect(err).NotTo(HaveOccurred())
					_, err = r.Reconcile(ctx, reconcile.Request{
						NamespacedName: typeNamespacedName,
					})
					Expect(err).NotTo(HaveOccurred())
					for _, expect := range expected {
						checkExist(expect)
					}
					By("removing the custom resource (not cleanup)")
					Eventually(func() error {
						return k8sClient.Delete(ctx, obj)
					}, timeout).Should(Succeed())
					_, err = r.Reconcile(ctx, reconcile.Request{
						NamespacedName: typeNamespacedName,
					})
					Expect(err).To(Not(HaveOccurred()))
					for _, expect := range expected {
						checkExist(expect)
					}
				},
			},
			{
				name: "EKS mode",
				obj: &irsav1alpha1.IRSASetup{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-resource-eks1",
						Namespace: "default",
					},
					Spec: irsav1alpha1.IRSASetupSpec{
						Cleanup:         false,
						Mode:            irsav1alpha1.ModeEks,
						IamOIDCProvider: "oidc.example",
					},
				},
				f: func(r *IRSASetupReconciler, obj *irsav1alpha1.IRSASetup) {
					typeNamespacedName := types.NamespacedName{
						Name:      obj.Name,
						Namespace: obj.Namespace,
					}
					_, err := r.Reconcile(ctx, reconcile.Request{
						NamespacedName: typeNamespacedName,
					})
					Expect(err).NotTo(HaveOccurred())
					_, err = r.Reconcile(ctx, reconcile.Request{
						NamespacedName: typeNamespacedName,
					})
					Expect(err).To(Not(HaveOccurred()))
					By("removing the custom resource (not cleanup)")
					Eventually(func() error {
						return k8sClient.Delete(ctx, obj)
					}, timeout).Should(Succeed())
					_, err = r.Reconcile(ctx, reconcile.Request{
						NamespacedName: typeNamespacedName,
					})
					Expect(err).To(Not(HaveOccurred()))
				},
			},
		}
		for _, tt := range tests {
			It(tt.name, func() {
				typeNamespacedName := types.NamespacedName{
					Name:      tt.obj.Name,
					Namespace: tt.obj.Namespace,
				}
				controllerReconciler := &IRSASetupReconciler{
					Client:    k8sClient,
					Scheme:    k8sClient.Scheme(),
					AwsClient: newMockAwsClient(&mockAwsIamAPI{}, &mockAwsS3API{}, &mockAwsStsAPI{}),
				}
				By("creating the custom resource for the Kind IRSASetup")
				err := k8sClient.Get(ctx, typeNamespacedName, tt.obj)
				if err != nil && errors.IsNotFound(err) {
					Expect(k8sClient.Create(ctx, tt.obj)).To(Succeed())
				}
				_, err = controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: typeNamespacedName,
				})
				Expect(err).NotTo(HaveOccurred())
				tt.f(controllerReconciler, tt.obj)
			})
		}
		BeforeEach(func() {
		})
		AfterEach(func() {
		})
	})
})

func newMockAwsClient(iam *mockAwsIamAPI, s3 *mockAwsS3API, sts *mockAwsStsAPI) awsclient.AwsClient {
	return &mockAwsClient{
		iam,
		s3,
		sts,
	}
}

type mockAwsClient struct {
	iam *mockAwsIamAPI
	s3  *mockAwsS3API
	sts *mockAwsStsAPI
}

func (m *mockAwsClient) IamClient() *awsclient.AwsIamClient {
	return &awsclient.AwsIamClient{Client: m.iam}
}

func (m *mockAwsClient) S3Client(region, bucketName string) *awsclient.AwsS3Client {
	return &awsclient.AwsS3Client{Client: m.s3}
}

func (m *mockAwsClient) StsClient() *awsclient.AwsStsClient {
	return &awsclient.AwsStsClient{Client: m.sts}
}

type (
	mockAwsIamAPI struct {
		createOidcErr                 error
		deleteOidcErr                 error
		createRoleErr                 error
		deleteRoleErr                 error
		updateAssumeRolePolicyError   error
		listAttachedRolePoliciesError error
		attachRolePolicyError         error
		detachRolePolicyError         error
	}
	mockAwsS3API struct {
		createBucketErr bool
		deleteBucketErr bool
	}
	mockAwsStsAPI struct{}
)

func (m *mockAwsIamAPI) CreateOpenIDConnectProvider(ctx context.Context, params *iam.CreateOpenIDConnectProviderInput, optFns ...func(*iam.Options)) (*iam.CreateOpenIDConnectProviderOutput, error) {
	return &iam.CreateOpenIDConnectProviderOutput{OpenIDConnectProviderArn: aws.String("arn::mock")}, m.createOidcErr
}

func (m *mockAwsIamAPI) DeleteOpenIDConnectProvider(ctx context.Context, params *iam.DeleteOpenIDConnectProviderInput, optFns ...func(*iam.Options)) (*iam.DeleteOpenIDConnectProviderOutput, error) {
	return &iam.DeleteOpenIDConnectProviderOutput{}, m.deleteOidcErr
}

func (m *mockAwsIamAPI) CreateRole(ctx context.Context, params *iam.CreateRoleInput, optFns ...func(*iam.Options)) (*iam.CreateRoleOutput, error) {
	return nil, m.createRoleErr
}

func (m *mockAwsIamAPI) ListAttachedRolePolicies(ctx context.Context, params *iam.ListAttachedRolePoliciesInput, optFns ...func(*iam.Options)) (*iam.ListAttachedRolePoliciesOutput, error) {
	return nil, m.listAttachedRolePoliciesError
}

func (m *mockAwsIamAPI) UpdateAssumeRolePolicy(ctx context.Context, params *iam.UpdateAssumeRolePolicyInput, optFns ...func(*iam.Options)) (*iam.UpdateAssumeRolePolicyOutput, error) {
	return nil, m.updateAssumeRolePolicyError
}

func (m *mockAwsIamAPI) AttachRolePolicy(ctx context.Context, params *iam.AttachRolePolicyInput, optFns ...func(*iam.Options)) (*iam.AttachRolePolicyOutput, error) {
	return nil, m.attachRolePolicyError
}

func (m *mockAwsIamAPI) DeleteRole(ctx context.Context, params *iam.DeleteRoleInput, optFns ...func(*iam.Options)) (*iam.DeleteRoleOutput, error) {
	return nil, m.deleteRoleErr
}

func (m *mockAwsIamAPI) DetachRolePolicy(ctx context.Context, params *iam.DetachRolePolicyInput, optFns ...func(*iam.Options)) (*iam.DetachRolePolicyOutput, error) {
	return nil, m.detachRolePolicyError
}

func (m *mockAwsStsAPI) GetCallerIdentity(ctx context.Context, params *sts.GetCallerIdentityInput, optFns ...func(*sts.Options)) (*sts.GetCallerIdentityOutput, error) {
	return &sts.GetCallerIdentityOutput{Account: aws.String("123456789012")}, nil
}

func (m *mockAwsS3API) CreateBucket(ctx context.Context, params *s3.CreateBucketInput, optFns ...func(*s3.Options)) (*s3.CreateBucketOutput, error) {
	if m.createBucketErr {
		return nil, fmt.Errorf("create bucket error")
	}
	return nil, nil
}

func (m *mockAwsS3API) PutObject(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error) {
	return nil, nil
}

func (m *mockAwsS3API) HeadObject(ctx context.Context, params *s3.HeadObjectInput, optFns ...func(*s3.Options)) (*s3.HeadObjectOutput, error) {
	return nil, &s3types.NotFound{}
}

func (m *mockAwsS3API) DeletePublicAccessBlock(ctx context.Context, params *s3.DeletePublicAccessBlockInput, optFns ...func(*s3.Options)) (*s3.DeletePublicAccessBlockOutput, error) {
	return nil, nil
}

func (m *mockAwsS3API) PutBucketOwnershipControls(ctx context.Context, params *s3.PutBucketOwnershipControlsInput, optFns ...func(*s3.Options)) (*s3.PutBucketOwnershipControlsOutput, error) {
	return nil, nil
}

func (m *mockAwsS3API) DeleteBucket(ctx context.Context, params *s3.DeleteBucketInput, optFns ...func(*s3.Options)) (*s3.DeleteBucketOutput, error) {
	if m.deleteBucketErr {
		return nil, fmt.Errorf("delete bucket error")
	}
	return nil, nil
}

func (m *mockAwsS3API) DeleteObjects(ctx context.Context, params *s3.DeleteObjectsInput, optFns ...func(*s3.Options)) (*s3.DeleteObjectsOutput, error) {
	return nil, nil
}
