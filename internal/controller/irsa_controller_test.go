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

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	irsav1alpha1 "github.com/kkb0318/irsa-manager/api/v1alpha1"
)

var _ = Describe("IRSA Controller", func() {
	Context("When reconciling IRSA", func() {
		tests := []struct {
			name         string
			obj          *irsav1alpha1.IRSA
			irsaSetupObj *irsav1alpha1.IRSASetup
			f            func(*IRSAReconciler, *irsav1alpha1.IRSA)
		}{
			{
				name: "should reconcile successfully",
				obj: &irsav1alpha1.IRSA{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-resource1",
						Namespace: "default",
					},
					Spec: irsav1alpha1.IRSASpec{
						Cleanup: true,
						ServiceAccount: irsav1alpha1.IRSAServiceAccount{
							Name: "sa-1",
							Namespaces: []string{
								"kube-system",
								"default",
							},
						},
					},
				},
				irsaSetupObj: newMockIRSASetup(),
				f: func(r *IRSAReconciler, obj *irsav1alpha1.IRSA) {
					expected := []expectedResource{
						{
							NamespacedName: types.NamespacedName{Name: "sa-1", Namespace: "kube-system"},
							f:              newServiceAccount,
						},
						{
							NamespacedName: types.NamespacedName{Name: "sa-1", Namespace: "default"},
							f:              newServiceAccount,
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
				name: "AWS API Error Case",
				obj: &irsav1alpha1.IRSA{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-resource2",
						Namespace: "default",
					},
					Spec: irsav1alpha1.IRSASpec{
						Cleanup: true,
						ServiceAccount: irsav1alpha1.IRSAServiceAccount{
							Name: "sa-2",
							Namespaces: []string{
								"kube-system",
								"default",
							},
						},
					},
				},
				irsaSetupObj: newMockIRSASetup(),
				f: func(r *IRSAReconciler, obj *irsav1alpha1.IRSA) {
					expected := []expectedResource{
						{
							NamespacedName: types.NamespacedName{Name: "sa-2", Namespace: "kube-system"},
							f:              newServiceAccount,
						},
						{
							NamespacedName: types.NamespacedName{Name: "sa-2", Namespace: "default"},
							f:              newServiceAccount,
						},
					}

					By("Reconciling the created resource")
					typeNamespacedName := types.NamespacedName{
						Name:      obj.Name,
						Namespace: obj.Namespace,
					}

					By("Error when creating role")
					r.AwsClient = newMockAwsClient(&mockAwsIamAPI{createRoleErr: fmt.Errorf("createRoleErr")}, nil, nil)
					_, err := r.Reconcile(ctx, reconcile.Request{
						NamespacedName: typeNamespacedName,
					})
					Expect(err).To(HaveOccurred())
					for _, expect := range expected {
						checkNoExist(expect)
					}

					By("successfully Reconciling")
					r.AwsClient = newMockAwsClient(&mockAwsIamAPI{}, nil, nil)
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
				name: "should update serviceaccount successfully",
				obj: &irsav1alpha1.IRSA{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-resource3",
						Namespace: "default",
					},
					Spec: irsav1alpha1.IRSASpec{
						Cleanup: true,
						ServiceAccount: irsav1alpha1.IRSAServiceAccount{
							Name: "sa-3",
							Namespaces: []string{
								"kube-system",
							},
						},
					},
				},
				irsaSetupObj: newMockIRSASetup(),
				f: func(r *IRSAReconciler, obj *irsav1alpha1.IRSA) {
					expected := []expectedResource{
						{
							NamespacedName: types.NamespacedName{Name: "sa-3", Namespace: "kube-system"},
							f:              newServiceAccount,
						},
						{
							NamespacedName: types.NamespacedName{Name: "sa-3", Namespace: "default"},
							f:              newServiceAccount,
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
					checkExist(
						expectedResource{
							NamespacedName: types.NamespacedName{Name: "sa-3", Namespace: "kube-system"},
							f:              newServiceAccount,
						},
					)
					f := createCallBackForFixingNamespace(ctx, r, typeNamespacedName, obj)

					By("Add Namespace 'default'")
					f(obj.Spec.ServiceAccount.Name, []string{"default", "kube-system"})
					By("Remove Namespace 'kube-system'")
					f(obj.Spec.ServiceAccount.Name, []string{"default"})
					checkNoExist(expectedResource{
						NamespacedName: types.NamespacedName{Name: "sa-3", Namespace: "kube-system"},
						f:              newServiceAccount,
					})

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
				name: "should update serviceaccount successfully with EKS mode",
				obj: &irsav1alpha1.IRSA{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-resource-eks-1",
						Namespace: "default",
					},
					Spec: irsav1alpha1.IRSASpec{
						Cleanup: true,
						ServiceAccount: irsav1alpha1.IRSAServiceAccount{
							Name: "sa-eks-1",
							Namespaces: []string{
								"kube-system",
							},
						},
					},
				},
				irsaSetupObj: newMockIRSASetupForEKS(),
				f: func(r *IRSAReconciler, obj *irsav1alpha1.IRSA) {
					expected := []expectedResource{
						{
							NamespacedName: types.NamespacedName{Name: "sa-eks-1", Namespace: "kube-system"},
							f:              newServiceAccount,
						},
						{
							NamespacedName: types.NamespacedName{Name: "sa-eks-1", Namespace: "default"},
							f:              newServiceAccount,
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
					checkExist(
						expectedResource{
							NamespacedName: types.NamespacedName{Name: "sa-eks-1", Namespace: "kube-system"},
							f:              newServiceAccount,
						},
					)
					f := createCallBackForFixingNamespace(ctx, r, typeNamespacedName, obj)

					By("Add Namespace 'default'")
					f(obj.Spec.ServiceAccount.Name, []string{"default", "kube-system"})
					By("Remove Namespace 'kube-system'")
					f(obj.Spec.ServiceAccount.Name, []string{"default"})
					checkNoExist(expectedResource{
						NamespacedName: types.NamespacedName{Name: "sa-eks-1", Namespace: "kube-system"},
						f:              newServiceAccount,
					})

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
		}
		for _, tt := range tests {
			It(tt.name, func() {
				typeNamespacedName := types.NamespacedName{
					Name:      tt.obj.Name,
					Namespace: tt.obj.Namespace,
				}
				controllerReconciler := &IRSAReconciler{
					Client:    k8sClient,
					Scheme:    k8sClient.Scheme(),
					AwsClient: newMockAwsClient(&mockAwsIamAPI{}, nil, nil),
				}
				By("creating the mock ISASetup")
				if tt.irsaSetupObj != nil {
					err := k8sClient.Get(ctx, client.ObjectKeyFromObject(tt.irsaSetupObj), &irsav1alpha1.IRSASetup{})
					if err != nil && errors.IsNotFound(err) {
						Expect(k8sClient.Create(ctx, tt.irsaSetupObj)).To(Succeed())
					}

				}
				By("creating the custom resource for the Kind IRSA")
				err := k8sClient.Get(ctx, typeNamespacedName, tt.obj)
				if err != nil && errors.IsNotFound(err) {
					Expect(k8sClient.Create(ctx, tt.obj)).To(Succeed())
				}
				_, err = controllerReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: typeNamespacedName,
				})
				Expect(err).NotTo(HaveOccurred())
				tt.f(controllerReconciler, tt.obj)

				By("deleting the mock ISASetup")
				err = k8sClient.Get(ctx, client.ObjectKeyFromObject(tt.irsaSetupObj), tt.irsaSetupObj)
				if err == nil {
					err = k8sClient.Delete(ctx, tt.irsaSetupObj)
					Expect(err).NotTo(HaveOccurred())
				}
			})
		}
		BeforeEach(func() {
		})
		AfterEach(func() {
		})
	})
})

func newMockIRSASetup() *irsav1alpha1.IRSASetup {
	return &irsav1alpha1.IRSASetup{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "default",
		},
		Spec: irsav1alpha1.IRSASetupSpec{
			Discovery: irsav1alpha1.Discovery{
				S3: irsav1alpha1.S3Discovery{
					Region:     "ap-northeast-1",
					BucketName: "irsa-manager-1",
				},
			},
		},
	}
}

func newMockIRSASetupForEKS() *irsav1alpha1.IRSASetup {
	return &irsav1alpha1.IRSASetup{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "default",
		},
		Spec: irsav1alpha1.IRSASetupSpec{
			Mode:            irsav1alpha1.ModeEks,
			IamOIDCProvider: "oidc.example",
		},
	}
}

func createCallBackForFixingNamespace(ctx context.Context, r *IRSAReconciler, typeNamespacedName types.NamespacedName, obj *irsav1alpha1.IRSA) func(name string, namespaces []string) {
	return func(name string, namespaces []string) {
		fixNamespacesAndReconcile(ctx, r, typeNamespacedName, obj, name, namespaces)
	}
}

func fixNamespacesAndReconcile(ctx context.Context, r *IRSAReconciler, typeNamespacedName types.NamespacedName, obj *irsav1alpha1.IRSA, name string, namespaces []string) {
	Expect(k8sClient.Get(ctx, typeNamespacedName, obj)).NotTo(HaveOccurred())
	Eventually(func() error {
		obj.Spec.ServiceAccount.Namespaces = namespaces
		return k8sClient.Update(ctx, obj)
	}, timeout).Should(Succeed())
	_, err := r.Reconcile(ctx, reconcile.Request{
		NamespacedName: typeNamespacedName,
	})
	Expect(err).NotTo(HaveOccurred())
	expected := []expectedResource{}
	for _, ns := range namespaces {
		expected = append(expected, expectedResource{
			NamespacedName: types.NamespacedName{Name: name, Namespace: ns},
			f:              newServiceAccount,
		})
	}
	for _, e := range expected {
		checkExist(e)
	}
}
