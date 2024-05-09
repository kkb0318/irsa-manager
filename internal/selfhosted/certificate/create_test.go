package certificate

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"testing"

	"k8s.io/apimachinery/pkg/types"
)

func TestCreateTlsCredentials(t *testing.T) {
	creds, err := CreateTlsCredential(types.NamespacedName{
		Name:      "pod-identity-webhook",
		Namespace: "kube-system",
	})
	if err != nil {
		t.Fatalf("Failed to create TLS credentials: %v", err)
	}

	certBlock, _ := pem.Decode(creds.certificate)
	if certBlock == nil {
		t.Fatal("Failed to decode PEM block containing the certificate")
	}
	cert, err := x509.ParseCertificate(certBlock.Bytes)
	if err != nil {
		t.Fatalf("Failed to parse certificate: %v", err)
	}

	keyBlock, _ := pem.Decode(creds.privateKey)
	if keyBlock == nil {
		t.Fatal("Failed to decode PEM block containing the private key")
	}
	key, err := x509.ParsePKCS1PrivateKey(keyBlock.Bytes)
	if err != nil {
		t.Fatalf("Failed to parse private key: %v", err)
	}

	// Verify public keys are equivalent
	if !publicKeysEqual(cert.PublicKey, &key.PublicKey) {
		t.Fatal("Public key in certificate does not match public key in private key")
	}
}

// Helper function to compare public keys
func publicKeysEqual(pub1, pub2 interface{}) bool {
	rsaPub1, ok1 := pub1.(*rsa.PublicKey)
	rsaPub2, ok2 := pub2.(*rsa.PublicKey)

	if !ok1 || !ok2 {
		return false
	}
	return rsaPub1.N.Cmp(rsaPub2.N) == 0 && rsaPub1.E == rsaPub2.E
}
