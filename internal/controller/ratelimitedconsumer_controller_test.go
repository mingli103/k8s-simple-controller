/*
Copyright 2025.

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
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"

	ratelimitv1alpha1 "github.com/mingli103/k8s-simple-controller/api/v1alpha1"
)

var _ = Describe("RateLimitedConsumer Controller", func() {
	Context("When reconciling a resource", func() {
		const (
			timeout  = time.Second * 10
			interval = time.Millisecond * 250
		)
		ctx := context.Background()

		var (
			routeName     = "test-route"
			pluginName    = []string{"test-rate-limit-5f2f1e04"}
			rlcName       = "test-rlc"
			testNamespace = "default"
		)
		typeNamespacedName := types.NamespacedName{
			Name:      rlcName,
			Namespace: "default",
		}
		ratelimitedconsumer := &ratelimitv1alpha1.RateLimitedConsumer{}

		BeforeEach(func() {
			// Create a test HTTPRoute
			By("creating the test HTTPRoute")
			route := &gatewayv1.HTTPRoute{
				ObjectMeta: metav1.ObjectMeta{
					Name:      routeName,
					Namespace: testNamespace,
				},
				Spec: gatewayv1.HTTPRouteSpec{
					Rules: []gatewayv1.HTTPRouteRule{
						{
							BackendRefs: []gatewayv1.HTTPBackendRef{
								{
									BackendRef: gatewayv1.BackendRef{
										BackendObjectReference: gatewayv1.BackendObjectReference{
											Name: gatewayv1.ObjectName("example-svc"),
											Port: func(p gatewayv1.PortNumber) *gatewayv1.PortNumber { return &p }(gatewayv1.PortNumber(80)),
										},
									},
								},
							},
						},
					},
				},
			}
			Expect(k8sClient.Create(ctx, route)).Should(Succeed())

			By("creating the custom resource for the Kind RateLimitedConsumer")
			err := k8sClient.Get(ctx, typeNamespacedName, ratelimitedconsumer)
			if err != nil && errors.IsNotFound(err) {
				resource := &ratelimitv1alpha1.RateLimitedConsumer{
					ObjectMeta: metav1.ObjectMeta{
						Name:      rlcName,
						Namespace: testNamespace,
					},
					Spec: ratelimitv1alpha1.RateLimitedConsumerSpec{
						RateLimit: ratelimitv1alpha1.RateLimit{
							Names: pluginName,
						},
						TargetRoute: ratelimitv1alpha1.TargetRoute{
							Name: routeName,
						},
					},
				}
				Expect(resource.Spec.RateLimit.Names).ToNot(BeEmpty())
				Expect(k8sClient.Create(ctx, resource)).To(Succeed())

			}
		})

		AfterEach(func() {
			// TODO(user): Cleanup logic after each test, like removing the resource instance.
			resource := &ratelimitv1alpha1.RateLimitedConsumer{}
			err := k8sClient.Get(ctx, typeNamespacedName, resource)
			Expect(err).NotTo(HaveOccurred())

			By("Cleanup the specific resource instance RateLimitedConsumer")
			Expect(k8sClient.Delete(ctx, resource)).To(Succeed())

			By("Cleanup the test HTTPRoute")
			Expect(k8sClient.Delete(ctx, &gatewayv1.HTTPRoute{
				ObjectMeta: metav1.ObjectMeta{
					Name:      routeName,
					Namespace: testNamespace,
				},
			})).To(Succeed())
		})
		It("should annotate the target HTTPRoute with the Kong plugin", func() {
			By("Reconciling the created resource")
			controllerReconciler := &RateLimitedConsumerReconciler{
				Client: k8sClient,
				Scheme: k8sClient.Scheme(),
			}

			_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())
			// Verify annotation on the route
			for _, expectedPlugin := range pluginName {
				Eventually(func(g Gomega) map[string]string {
					var updated gatewayv1.HTTPRoute
					err := k8sClient.Get(ctx, types.NamespacedName{Name: routeName, Namespace: testNamespace}, &updated)
					g.Expect(err).ToNot(HaveOccurred())
					fmt.Println("Current annotation value:", updated.Annotations["konghq.com/plugins"], "expected:", expectedPlugin)
					return updated.Annotations
				}, timeout, interval).Should(HaveKeyWithValue("konghq.com/plugins", expectedPlugin))
			}
		})
	})
})
