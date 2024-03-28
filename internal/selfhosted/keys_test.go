package selfhosted

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReadKey(t *testing.T) {
	t.Run("key pair check", func(t *testing.T) {
		keyPair, err := createKeyPair()
		assert.NoError(t, err)

		message := []byte("test message")
		hashed := sha256.Sum256(message)

		block, _ := pem.Decode(keyPair.PrivateKey)
		assert.NotNil(t, block, "failed to decode private key to PEM")

		privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
		assert.NoError(t, err, "failed to parse private key")

		signature, err := rsa.SignPKCS1v15(rand.Reader, privateKey, crypto.SHA256, hashed[:])

		assert.NoError(t, err, "failed to create signature")
		block, _ = pem.Decode(keyPair.PublicKey)
		assert.NotNil(t, block, "failed to decode public key to PEM")

		pubKey, err := x509.ParsePKIXPublicKey(block.Bytes)
		assert.NoError(t, err, "failed to parse public key")

		rsaPubKey, ok := pubKey.(*rsa.PublicKey)
		assert.Truef(t, ok, "public key is not RSA")

		err = rsa.VerifyPKCS1v15(rsaPubKey, crypto.SHA256, hashed[:], signature)
		assert.NoError(t, err, "failed to check signature")
	})
}
