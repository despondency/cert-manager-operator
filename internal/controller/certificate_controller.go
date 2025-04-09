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

	"github.com/despondency/cert-manager-operator/internal/cert"
	v1core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	certsk8ciov1 "github.com/despondency/cert-manager-operator/api/v1"
)

// CertificateReconciler reconciles a Certificate object
type CertificateReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=certs.k8c.io.despondency.io,resources=certificates,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=certs.k8c.io.despondency.io,resources=certificates/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=certs.k8c.io.despondency.io,resources=certificates/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Certificate object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.20.4/pkg/reconcile
func (r *CertificateReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := logf.FromContext(ctx)
	c := &certsk8ciov1.Certificate{}
	if err := r.Get(ctx, req.NamespacedName, c); err != nil {
		logger.Error(err, "unable to fetch Certificate")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	// get the secretRef
	s := &v1core.Secret{}
	if err := r.Client.Get(ctx,
		client.ObjectKey{
			Namespace: req.Namespace,
			Name:      c.Spec.SecretRef},
		s); err != nil {
		if errors.IsNotFound(err) {
			// let's pretend that the cert provisioning "might take" more time than usual
			// so we put this intermediate status
			c.Status.Status = certsk8ciov1.StatusProvisioning
			err := r.Status().Update(ctx, c)
			if err != nil {
				return ctrl.Result{}, err
			}
			// if the secret is not found, we need to create it
			// create the certificate
			certDetails, err := cert.CreateCertificate(c.Spec.DNSName, c.Spec.Validity)
			if err != nil {
				return ctrl.Result{}, err
			}
			s.Name = c.Spec.SecretRef
			s.Namespace = req.Namespace
			s.Data = map[string][]byte{
				"ca.crt":           certDetails.CaCert,
				"tls.key":          certDetails.TLSKey,
				"tls.crt":          certDetails.TLSCert,
				"tls-combined.pem": certDetails.TLSCombined,
				"key.der":          certDetails.KeyBase64,
			}
			err = r.Client.Create(ctx, s)
			if err != nil {
				return ctrl.Result{}, err
			}
		}
	}
	// if secret is found
	// check the validity of the certificate
	// if the certificate is going to expire (based on some threshold)
	// regenerate the cert + secret

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *CertificateReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&certsk8ciov1.Certificate{}).
		Named("certificate").
		Owns(&v1core.Secret{}).
		Complete(r)
}
