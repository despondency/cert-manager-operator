package cert

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"strconv"
	"strings"
	"time"
)

func parseDuration(s string) (time.Duration, error) {
	if strings.HasSuffix(s, "d") {
		numStr := strings.TrimSuffix(s, "d")
		days, err := strconv.Atoi(numStr)
		if err != nil {
			return 0, err
		}
		return time.Duration(days) * 24 * time.Hour, nil
	} else {
		return -1, fmt.Errorf("invalid duration: %s, only days supported", s)
	}
}

type X509Cert struct {
	CaCert      []byte
	TLSKey      []byte
	TLSCert     []byte
	TLSCombined []byte
	KeyBase64   []byte
}

func CreateCertificate(host string, validity string) (*X509Cert, error) {

	// === 1. Generate CA key ===
	caKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, err
	}

	certDurationInDays, err := parseDuration(validity)
	if err != nil {
		return nil, err
	}

	notBefore := time.Now()
	notAfter := notBefore.Add(certDurationInDays)

	// === 2. Self-signed CA certificate ===
	caTemplate := &x509.Certificate{
		SerialNumber:          big.NewInt(2024),
		Subject:               pkix.Name{CommonName: "My Organization CA"},
		NotBefore:             notBefore,
		NotAfter:              notAfter,
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageDigitalSignature,
		BasicConstraintsValid: true,
		IsCA:                  true,
	}

	caDER, err := x509.CreateCertificate(rand.Reader, caTemplate, caTemplate, &caKey.PublicKey, caKey)
	if err != nil {
		return nil, err
	}

	// === 3. Generate TLS key ===
	tlsKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, err
	}

	// === 4. TLS certificate signed by CA ===
	tlsTemplate := &x509.Certificate{
		SerialNumber: big.NewInt(1001),
		Subject: pkix.Name{
			CommonName:   "example.com",
			Organization: []string{"Example Org"},
		},
		NotBefore:             notBefore,
		NotAfter:              notAfter,
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	hosts := strings.Split(host, ",")
	for _, h := range hosts {
		if ip := net.ParseIP(h); ip != nil {
			tlsTemplate.IPAddresses = append(tlsTemplate.IPAddresses, ip)
		} else {
			tlsTemplate.DNSNames = append(tlsTemplate.DNSNames, h)
		}
	}

	tlsCertDER, err := x509.CreateCertificate(rand.Reader, tlsTemplate, caTemplate, &tlsKey.PublicKey, caKey)
	if err != nil {
		return nil, err
	}

	// === 5. PEM Encodings ===
	caPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: caDER})
	tlsKeyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(tlsKey)})
	tlsCertPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: tlsCertDER})

	tlsCombinedPEM := append(tlsKeyPEM, tlsCertPEM...)
	tlsKeyDER := x509.MarshalPKCS1PrivateKey(tlsKey)

	return &X509Cert{
		CaCert:      caPEM,
		TLSKey:      tlsKeyPEM,
		TLSCert:     tlsCertPEM,
		TLSCombined: tlsCombinedPEM,
		KeyBase64:   tlsKeyDER,
	}, nil
}
