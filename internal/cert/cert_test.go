package cert_test

import (
	"crypto/x509"
	"encoding/pem"
	"testing"
	"time"

	"github.com/despondency/cert-manager-operator/internal/cert"
)

func TestCreateCert(t *testing.T) {
	c, err := cert.CreateCertificate("my-domain.io", "365d")
	// Parse CA cert
	if err != nil {
		t.Fatal(err)
	}
	caBlock, _ := pem.Decode(c.CaCert)
	if caBlock == nil {
		t.Fatal("failed to parse CA certificate PEM")
	}
	caCert, err := x509.ParseCertificate(caBlock.Bytes)
	if err != nil {
		t.Fatal(err)
	}

	// Parse TLS cert
	tlsBlock, _ := pem.Decode(c.TLSCert)
	if tlsBlock == nil {
		t.Fatal("failed to parse TLS certificate PEM")
	}
	tlsCert, err := x509.ParseCertificate(tlsBlock.Bytes)
	if err != nil {
		t.Fatal(err)
	}

	// Check validity period
	now := time.Now()
	if now.Before(tlsCert.NotBefore) {
		t.Errorf("certificate is not valid yet: starts at %v", tlsCert.NotBefore)
	}
	if now.After(tlsCert.NotAfter) {
		t.Errorf("certificate has expired: expired at %v", tlsCert.NotAfter)
	}
	if tlsCert.NotAfter.Sub(tlsCert.NotBefore).Hours()/24 != 365 {
		t.Errorf("certificate validity is not 365 days")
	}
	t.Logf("certificate is valid from %v to %v", tlsCert.NotBefore, tlsCert.NotAfter)

	// Verify with SAN domain
	roots := x509.NewCertPool()
	roots.AddCert(caCert)

	opts := x509.VerifyOptions{
		Roots:   roots,
		DNSName: "my-domain.io", // 🔍 check against SAN
	}

	if _, err := tlsCert.Verify(opts); err != nil {
		t.Errorf("certificate validation failed: %v", err)
	} else {
		t.Logf("certificate is valid for SAN: %s", "my-domain.io")
	}
}
