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
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	c "sigs.k8s.io/controller-runtime/pkg/client"

	certsk8ciov1 "github.com/despondency/cert-manager-operator/api/v1"
)

const (
	timeout  = time.Second * 10
	interval = time.Second * 1
)

var _ = Describe("Certificate Controller", func() {
	Context("When creating a Certificate", func() {
		It("should successfully create the certificate", func() {
			resource := &certsk8ciov1.Certificate{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "new-cert",
					Namespace: "default",
				},
				Spec: certsk8ciov1.CertificateSpec{
					DNSName:   "io.despondency.com",
					Validity:  "365d",
					SecretRef: "certificate-secret",
				},
			}
			Expect(k8sClient.Create(ctx, resource)).To(Succeed())

			Eventually(func(g Gomega) {
				g.Expect(k8sClient.Get(ctx, types.NamespacedName{
					Name:      "new-cert",
					Namespace: "default",
				}, resource)).To(Succeed())
				g.Expect(resource.Status.Status).To(Equal(certsk8ciov1.StatusProvisioning))
			}, timeout, interval).Should(Succeed())

			certSecret := &v1.Secret{}
			Eventually(func(g Gomega) {
				g.Expect(k8sClient.Get(ctx, types.NamespacedName{
					Name:      "certificate-secret",
					Namespace: "default",
				}, certSecret)).To(Succeed())
			}, timeout, interval).Should(Succeed())

			// Delete the parent resource (MyKind)
			err := k8sClient.Delete(ctx, resource)
			Expect(err).NotTo(HaveOccurred())

			Eventually(func(g Gomega) {
				g.Expect(k8sClient.Get(ctx, c.ObjectKey{Name: "new-cert", Namespace: "default"}, resource)).To(Succeed())
			}, timeout, interval).ShouldNot(Succeed())

			Eventually(func(g Gomega) {
				secret := &v1.Secret{}
				g.Expect(k8sClient.Get(ctx, c.ObjectKey{Name: "certificate-secret", Namespace: "default"}, secret)).To(Succeed())
			}, timeout, interval).ShouldNot(Succeed())

		})
	})
})
