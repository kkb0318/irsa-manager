package oidc

import (
	"encoding/json"
	"fmt"

	"github.com/kkb0318/irsa-manager/internal/selfhosted"
)

type MyIdProvider struct {
	keyPair *selfhosted.KeyPair
	jwk     *selfhosted.JWK
}

func (p *MyIdProvider) Discovery() []byte {
	return []byte{}
}

func (p *MyIdProvider) JWK() ([]byte, error) {
	jsonData, err := json.MarshalIndent(p.jwk.GetKeys(), "", "  ")
	if err != nil {
		return nil, fmt.Errorf("error marshalling JSON: %s", err.Error())
	}
	return jsonData, nil
}

func (p *MyIdProvider) Endpoint() []byte {
	return []byte{}
}

func NewMyIdProvider(keyPair *selfhosted.KeyPair, jwk *selfhosted.JWK) *MyIdProvider {
	return &MyIdProvider{keyPair, jwk}
}
