package selfhosted

import (
	"crypto"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"errors"
	"fmt"

	jose "github.com/go-jose/go-jose/v4"
	"k8s.io/client-go/util/keyutil"
)

// keyIDFromPublicKey derives a key ID non-reversibly from a public key.
//
// The Key ID is field on a given on JWTs and JWKs that help relying parties
// pick the correct key for verification when the identity party advertises
// multiple keys.
//
// Making the derivation non-reversible makes it impossible for someone to
// accidentally obtain the real key from the key ID and use it for token
// validation.
// This method is copied from
// https://github.com/kubernetes/kubernetes/blob/v1.29.3/pkg/serviceaccount/jwt.go#L99
func keyIDFromPublicKey(publicKey interface{}) (string, error) {
	publicKeyDERBytes, err := x509.MarshalPKIXPublicKey(publicKey)
	if err != nil {
		return "", fmt.Errorf("failed to serialize public key to DER format: %v", err)
	}

	hasher := crypto.SHA256.New()
	hasher.Write(publicKeyDERBytes)
	publicKeyDERHash := hasher.Sum(nil)

	keyID := base64.RawURLEncoding.EncodeToString(publicKeyDERHash)

	return keyID, nil
}

type JWK struct {
	Keys []jose.JSONWebKey `json:"keys"`
}

func NewJWK(pub []byte) (*JWK, error) {
	pubKeys, err := keyutil.ParsePublicKeysPEM(pub)
	if err != nil {
		return nil, err
	}
	pubKey := pubKeys[0]
	var alg jose.SignatureAlgorithm
	switch pubKey.(type) {
	case *rsa.PublicKey:
		alg = jose.RS256
	default:
		return nil, errors.New("public key is not RSA")
	}

	kid, err := keyIDFromPublicKey(pubKey)
	if err != nil {
		return nil, err
	}

	var keys []jose.JSONWebKey
	keys = append(keys, jose.JSONWebKey{
		Key:       pubKey,
		KeyID:     kid,
		Algorithm: string(alg),
		Use:       "sig",
	})
	keys = append(keys, jose.JSONWebKey{
		Key:       pubKey,
		KeyID:     "",
		Algorithm: string(alg),
		Use:       "sig",
	})
	return &JWK{Keys: keys}, nil
}
