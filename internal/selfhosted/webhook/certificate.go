package webhook

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"time"

	"k8s.io/apimachinery/pkg/types"
)

type TlsCredential struct {
	privateKey  []byte
	certificate []byte
}

func (t TlsCredential) Certificate() []byte {
	return t.certificate
}

func (t TlsCredential) PrivateKey() []byte {
	return t.privateKey
}

func CreateTlsCredential(serviceNamespacedName types.NamespacedName) (TlsCredential, error) {
	certificatePeriod := 365 // days

	// Generate RSA private key
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return TlsCredential{}, err
	}

	// Define certificate template
	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			CommonName: serviceNamespacedName.Name + "." + serviceNamespacedName.Namespace + ".svc",
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(0, 0, certificatePeriod),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		IsCA:                  true,
	}

	// Add SANs to the certificate template
	template.DNSNames = []string{
		serviceNamespacedName.Name + "." + serviceNamespacedName.Namespace + ".svc",
		serviceNamespacedName.Name + "." + serviceNamespacedName.Namespace + ".svc.cluster.local",
	}

	// Create the certificate
	certBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	if err != nil {
		return TlsCredential{}, err
	}

	// Encode the private key to PEM format
	privPemBytes := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	})

	// Encode the certificate to PEM format
	certPemBytes := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certBytes,
	})

	return TlsCredential{privateKey: privPemBytes, certificate: certPemBytes}, nil
}
