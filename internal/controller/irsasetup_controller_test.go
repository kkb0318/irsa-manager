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

	awsclient "github.com/kkb0318/irsa-manager/internal/client"
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
				name: "case1",
				obj: &irsav1alpha1.IRSASetup{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-resource11",
						Namespace: "default",
					},
					Spec: irsav1alpha1.IRSASetupSpec{
						Mode: "selfhoted",
						Discovery: irsav1alpha1.Discovery{
							S3: irsav1alpha1.S3Discovery{
								Region:     "ap-northeast-1",
								BucketName: "irsa-manager-kkb-1",
							},
						},
					},
				},
				f: func(r *IRSASetupReconciler, obj *irsav1alpha1.IRSASetup) {
					expected := []types.NamespacedName{
						{Name: "irsa-manager-key", Namespace: "kube-system"},
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
						checkExist(expect, newSecret)
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
						checkNoExist(expect, newSecret)
					}
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
					resource := &irsav1alpha1.IRSASetup{
						ObjectMeta: metav1.ObjectMeta{
							Name:      typeNamespacedName.Name,
							Namespace: typeNamespacedName.Namespace,
						},
						Spec: irsav1alpha1.IRSASetupSpec{
							Mode: "selfhoted",
							Discovery: irsav1alpha1.Discovery{
								S3: irsav1alpha1.S3Discovery{
									Region:     "ap-northeast-1",
									BucketName: "irsa-manager-kkb-1",
								},
							},
						},
					}
					Expect(k8sClient.Create(ctx, resource)).To(Succeed())
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
		It("should successfully reconcile the resource", func() {
			const resourceName = "test-resource"

			typeNamespacedName := types.NamespacedName{
				Name:      resourceName,
				Namespace: "default",
			}
			irsasetup := &irsav1alpha1.IRSASetup{
				ObjectMeta: metav1.ObjectMeta{
					Name:      resourceName,
					Namespace: "default",
				},
			}
			By("creating the custom resource for the Kind IRSASetup")
			err := k8sClient.Get(ctx, typeNamespacedName, irsasetup)
			if err != nil && errors.IsNotFound(err) {
				resource := &irsav1alpha1.IRSASetup{
					ObjectMeta: metav1.ObjectMeta{
						Name:      resourceName,
						Namespace: "default",
					},
					Spec: irsav1alpha1.IRSASetupSpec{
						Mode: "selfhoted",
						Discovery: irsav1alpha1.Discovery{
							S3: irsav1alpha1.S3Discovery{
								Region:     "ap-northeast-1",
								BucketName: "irsa-manager-kkb-1",
							},
						},
					},
				}
				Expect(k8sClient.Create(ctx, resource)).To(Succeed())
			}

			expected := []types.NamespacedName{
				{Name: "irsa-manager-key", Namespace: "kube-system"},
			}

			By("Reconciling the created resource")
			controllerReconciler := &IRSASetupReconciler{
				Client:    k8sClient,
				Scheme:    k8sClient.Scheme(),
				AwsClient: newMockAwsClient(&mockAwsIamAPI{}, &mockAwsS3API{}, &mockAwsStsAPI{}),
			}

			_, err = controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())
			_, err = controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())
			for _, expect := range expected {
				checkExist(expect, newSecret)
			}
			By("removing the custom resource for the Kind")
			Eventually(func() error {
				return k8sClient.Delete(ctx, irsasetup)
			}, timeout).Should(Succeed())
			_, err = controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).To(Not(HaveOccurred()))
			for _, expect := range expected {
				checkNoExist(expect, newSecret)
			}
		})

		It("should successfully reconcile the resource", func() {
			const resourceName = "test-resource2"

			typeNamespacedName := types.NamespacedName{
				Name:      resourceName,
				Namespace: "default",
			}
			irsasetup := &irsav1alpha1.IRSASetup{
				ObjectMeta: metav1.ObjectMeta{
					Name:      resourceName,
					Namespace: "default",
				},
			}
			By("creating the custom resource for the Kind IRSASetup")
			err := k8sClient.Get(ctx, typeNamespacedName, irsasetup)
			if err != nil && errors.IsNotFound(err) {
				resource := &irsav1alpha1.IRSASetup{
					ObjectMeta: metav1.ObjectMeta{
						Name:      resourceName,
						Namespace: "default",
					},
					Spec: irsav1alpha1.IRSASetupSpec{
						Mode: "selfhoted",
						Discovery: irsav1alpha1.Discovery{
							S3: irsav1alpha1.S3Discovery{
								Region:     "ap-northeast-1",
								BucketName: "irsa-manager-kkb-1",
							},
						},
					},
				}
				Expect(k8sClient.Create(ctx, resource)).To(Succeed())
			}
			expected := []types.NamespacedName{
				{Name: "irsa-manager-key", Namespace: "kube-system"},
			}

			By("Reconciling the created resource")
			controllerReconciler := &IRSASetupReconciler{
				Client:    k8sClient,
				Scheme:    k8sClient.Scheme(),
				AwsClient: newMockAwsClient(&mockAwsIamAPI{}, &mockAwsS3API{}, &mockAwsStsAPI{}),
			}

			_, err = controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())
			By("Reconciling with the AwsClient error")
			controllerReconciler.AwsClient = newMockAwsClient(&mockAwsIamAPI{createOidcErr: true}, &mockAwsS3API{}, &mockAwsStsAPI{})
			_, err = controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).To(HaveOccurred())
			for _, expect := range expected {
				checkNoExist(expect, newSecret)
			}
			By("Reconciling successfully")
			controllerReconciler.AwsClient = newMockAwsClient(&mockAwsIamAPI{}, &mockAwsS3API{}, &mockAwsStsAPI{})
			_, err = controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())
			for _, expect := range expected {
				checkExist(expect, newSecret)
			}
			By("removing the custom resource for the Kind")
			Eventually(func() error {
				return k8sClient.Delete(ctx, irsasetup)
			}, timeout).Should(Succeed())
			_, err = controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).To(Not(HaveOccurred()))
			for _, expect := range expected {
				checkNoExist(expect, newSecret)
			}
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
		createOidcErr bool
		deleteOidcErr bool
	}
	mockAwsS3API struct {
		createBucketErr bool
		deleteBucketErr bool
	}
	mockAwsStsAPI struct {
		isErr bool
	}
)

func (m *mockAwsIamAPI) CreateOpenIDConnectProvider(ctx context.Context, params *iam.CreateOpenIDConnectProviderInput, optFns ...func(*iam.Options)) (*iam.CreateOpenIDConnectProviderOutput, error) {
	if m.createOidcErr {
		return nil, fmt.Errorf("create Oidc error")
	}
	return &iam.CreateOpenIDConnectProviderOutput{OpenIDConnectProviderArn: aws.String("arn::mock")}, nil
}

func (m *mockAwsIamAPI) DeleteOpenIDConnectProvider(ctx context.Context, params *iam.DeleteOpenIDConnectProviderInput, optFns ...func(*iam.Options)) (*iam.DeleteOpenIDConnectProviderOutput, error) {
	if m.deleteOidcErr {
		return nil, fmt.Errorf("delete Oidc error")
	}
	return &iam.DeleteOpenIDConnectProviderOutput{}, nil
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
