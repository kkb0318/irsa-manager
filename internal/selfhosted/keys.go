package selfhosted

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
)

type KeyPair struct {
	publicKey  []byte
	privateKey []byte
}

func CreateKeyPair() (*KeyPair, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, err
	}

	// convert private key to PEM
	privPem := pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	}
	privPemBytes := pem.EncodeToMemory(&privPem)

	// convert public key to PKIX, ASN.1 DER
	pubASN1, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		return nil, err
	}

	pubPem := pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: pubASN1,
	}
	pubPemBytes := pem.EncodeToMemory(&pubPem)
	return &KeyPair{pubPemBytes, privPemBytes}, nil
}

func (k *KeyPair) PublicKey() []byte {
	return k.publicKey
}

func (k *KeyPair) PrivateKey() []byte {
	return k.privateKey
}
