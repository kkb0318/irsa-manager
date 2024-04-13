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

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/s3"
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

		BeforeEach(func() {
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
		})

		It("should successfully reconcile the resource", func() {
			awsClient := newMockAwsClient()
			expected := []types.NamespacedName{
				// TODO:
				{Name: "name", Namespace: "default"},
			}

			By("Reconciling the created resource")
			controllerReconciler := &IRSASetupReconciler{
				Client:    k8sClient,
				Scheme:    k8sClient.Scheme(),
				AwsClient: awsClient,
			}

			_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
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
	})
})

func newMockAwsClient() awsclient.AwsClient {
	return &mockAwsClient{}
}

type mockAwsClient struct{}

func (m *mockAwsClient) IamClient() *awsclient.AwsIamClient {
	return &awsclient.AwsIamClient{Client: &mockAwsIamAPI{}}
}

func (m *mockAwsClient) S3Client(region, bucketName string) *awsclient.AwsS3Client {
	return &awsclient.AwsS3Client{Client: &mockAwsS3API{}}
}

type (
	mockAwsIamAPI struct{}
	mockAwsS3API  struct{}
)

func (m *mockAwsIamAPI) CreateOpenIDConnectProvider(ctx context.Context, params *iam.CreateOpenIDConnectProviderInput, optFns ...func(*iam.Options)) (*iam.CreateOpenIDConnectProviderOutput, error) {
	return &iam.CreateOpenIDConnectProviderOutput{OpenIDConnectProviderArn: aws.String("arn::mock")}, nil
}

func (m *mockAwsS3API) CreateBucket(ctx context.Context, params *s3.CreateBucketInput, optFns ...func(*s3.Options)) (*s3.CreateBucketOutput, error) {
	return nil, nil
}

func (m *mockAwsS3API) PutObject(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error) {
	return nil, nil
}

func (m *mockAwsS3API) DeletePublicAccessBlock(ctx context.Context, params *s3.DeletePublicAccessBlockInput, optFns ...func(*s3.Options)) (*s3.DeletePublicAccessBlockOutput, error) {
	return nil, nil
}

func (m *mockAwsS3API) PutBucketOwnershipControls(ctx context.Context, params *s3.PutBucketOwnershipControlsInput, optFns ...func(*s3.Options)) (*s3.PutBucketOwnershipControlsOutput, error) {
	return nil, nil
}

func (m *mockAwsS3API) DeleteBucket(ctx context.Context, params *s3.DeleteBucketInput, optFns ...func(*s3.Options)) (*s3.DeleteBucketOutput, error) {
	return nil, nil
}

func (m *mockAwsS3API) DeleteObjects(ctx context.Context, params *s3.DeleteObjectsInput, optFns ...func(*s3.Options)) (*s3.DeleteObjectsOutput, error) {
	return nil, nil
}
