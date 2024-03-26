package selfhosted

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"os"
)

func createKeyPair() error {
	// RSAキーペアの生成
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return err
	}

	// private keyをPEM形式で保存
	privPem := pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	}
	privPemFile, err := os.Create("private_key.pem")
	if err != nil {
		return err
	}
	defer privPemFile.Close()
	if err := pem.Encode(privPemFile, &privPem); err != nil {
		return err
	}

	// 公開鍵をPKIX, ASN.1 DER形式に変換
	pubASN1, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		return err
	}

	// 公開鍵をPEM形式で保存
	pubPem := pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: pubASN1,
	}
	pubPemFile, err := os.Create("public_key.pem")
	if err != nil {
		return err
	}
	defer pubPemFile.Close()
	if err := pem.Encode(pubPemFile, &pubPem); err != nil {
		return err
	}
	return nil
}
