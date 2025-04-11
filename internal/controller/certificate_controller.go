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
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"log"

	"github.com/despondency/cert-manager-operator/internal/cert"
	v1core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	certsk8ciov1 "github.com/despondency/cert-manager-operator/api/v1"
)

// CertificateReconciler reconciles a Certificate object
type CertificateReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

const (
	CACert      = "ca.crt"
	TLSKey      = "tls.key"
	TLSCert     = "tls.crt"
	TLSCombined = "tls-combined.pem"
	Key         = "key.der"
)

// +kubebuilder:rbac:groups=certs.k8c.io.despondency.io,resources=certificates,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=certs.k8c.io.despondency.io,resources=certificates/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=certs.k8c.io.despondency.io,resources=certificates/finalizers,verbs=update
// +kubebuilder:rbac:groups=core,resources=secrets,verbs=get;list;watch;create;update;patch;delete

func (r *CertificateReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = logf.FromContext(ctx)
	c := &certsk8ciov1.Certificate{}
	if err := r.Get(ctx, req.NamespacedName, c); err != nil {
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
			// if the secret is not found, we need to create it
			// create the certificate
			err = r.createCertificate(ctx, c)
			if err != nil {
				return ctrl.Result{}, err
			}
		}
	} else {
		// if secret is found
		// check the validity of the certificate
		// regenerate the cert + secret
		err := validateCertificate(s)
		if err != nil {
			// regenerate the certificate
			err = r.Client.Delete(ctx, s)
			if err != nil {
				return ctrl.Result{}, err
			}
			err = r.createCertificate(ctx, c)
			if err != nil {
				return ctrl.Result{}, err
			}
		}
	}

	return ctrl.Result{}, nil
}

func (r *CertificateReconciler) createCertificate(ctx context.Context,
	c *certsk8ciov1.Certificate) error {
	certSecret := &v1core.Secret{}
	certDetails, err := cert.CreateCertificate(c.Spec.DNSName, c.Spec.Validity)
	if err != nil {
		return err
	}
	certSecret.Name = c.Spec.SecretRef
	certSecret.Namespace = c.Namespace
	certSecret.Data = map[string][]byte{
		CACert:      certDetails.CaCert,
		TLSKey:      certDetails.TLSKey,
		TLSCert:     certDetails.TLSCert,
		TLSCombined: certDetails.TLSCombined,
		Key:         certDetails.KeyBase64,
	}
	err = controllerutil.SetControllerReference(c, certSecret, r.Scheme)
	if err != nil {
		return err
	}
	err = r.Client.Create(ctx, certSecret)
	if err != nil {
		return err
	}
	c.Status.Status = certsk8ciov1.StatusReady
	err = r.Status().Update(ctx, c)
	if err != nil {
		return err
	}
	return nil
}

func validateCertificate(s *v1core.Secret) error {
	caPEM, ok := s.Data[CACert]
	if !ok {
		return fmt.Errorf("secret %s/%s does not contain a certificate", s.Namespace, s.Name)
	}
	keyPEM, ok := s.Data[TLSKey]
	if !ok {
		return fmt.Errorf("secret %s/%s does not contain a certificate", s.Namespace, s.Name)
	}
	certPEM, ok := s.Data[TLSCert]
	if !ok {
		return fmt.Errorf("secret %s/%s does not contain a certificate", s.Namespace, s.Name)
	}
	// Parse certificate and key
	_, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		return fmt.Errorf("failed to parse certificate/key pair: %v", err)
	}
	// Load leaf certificate
	block, _ := pem.Decode(certPEM)
	if block == nil {
		return fmt.Errorf("failed to decode PEM block from certificate")
	}
	leafCert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return fmt.Errorf("failed to parse certificate: %v", err)
	}
	// Load CA cert pool
	roots := x509.NewCertPool()
	if ok := roots.AppendCertsFromPEM(caPEM); !ok {
		log.Fatalf("failed to parse CA cert")
	}
	// Verify certificate
	opts := x509.VerifyOptions{
		Roots: roots,
	}
	if _, err := leafCert.Verify(opts); err != nil {
		log.Fatalf("failed to verify certificate: %v", err)
	}
	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *CertificateReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&certsk8ciov1.Certificate{}).
		Named("certificate").
		Owns(&v1core.Secret{}).
		Complete(r)
}
