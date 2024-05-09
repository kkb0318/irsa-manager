package certificate

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/pem"
	"math/big"
	"time"

	"k8s.io/apimachinery/pkg/types"
)

type TlsCredentials struct {
	privateKey  []byte
	certificate []byte
}

func (t *TlsCredentials) CaBundle() string {
	return base64.StdEncoding.EncodeToString(t.certificate)
}

func (t *TlsCredentials) Certificate() []byte {
	return t.certificate
}

func (t *TlsCredentials) PrivateKey() []byte {
	return t.privateKey
}

func CreateTlsCredential(serviceNamespacedName types.NamespacedName) (TlsCredentials, error) {
	certificatePeriod := 365 // days

	// Generate RSA private key
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return TlsCredentials{}, err
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

	// Create the certificate
	certBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	if err != nil {
		return TlsCredentials{}, err
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

	return TlsCredentials{privateKey: privPemBytes, certificate: certPemBytes}, nil
}
