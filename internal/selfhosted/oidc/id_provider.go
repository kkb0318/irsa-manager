package oidc

import (
	"encoding/json"
	"fmt"

	"github.com/kkb0318/irsa-manager/internal/selfhosted"
)

type MyIdProvider struct {
	jwk            *selfhosted.JWK
	issuerHostPath string
}

type OIDCDiscoveryConfiguration struct {
	Issuer                           string   `json:"issuer"`
	JWKSURI                          string   `json:"jwks_uri"`
	AuthorizationEndpoint            string   `json:"authorization_endpoint"`
	ResponseTypesSupported           []string `json:"response_types_supported"`
	SubjectTypesSupported            []string `json:"subject_types_supported"`
	IDTokenSigningAlgValuesSupported []string `json:"id_token_signing_alg_values_supported"`
	ClaimsSupported                  []string `json:"claims_supported"`
}

func (p *MyIdProvider) Discovery() ([]byte, error) {
	oidcConfig := OIDCDiscoveryConfiguration{
		Issuer:                           fmt.Sprintf("https://%s/", p.issuerHostPath),
		JWKSURI:                          fmt.Sprintf("https://%s/keys.json", p.issuerHostPath),
		AuthorizationEndpoint:            "urn:kubernetes:programmatic_authorization",
		ResponseTypesSupported:           []string{"id_token"},
		SubjectTypesSupported:            []string{"public"},
		IDTokenSigningAlgValuesSupported: []string{"RS256"},
		ClaimsSupported:                  []string{"sub", "iss"},
	}
	jsonData, err := json.MarshalIndent(oidcConfig, "", "    ")
	if err != nil {
		return nil, err
	}
	return jsonData, nil
}

func (p *MyIdProvider) JWK() ([]byte, error) {
	jsonData, err := json.MarshalIndent(p.jwk.GetKeys(), "", "  ")
	if err != nil {
		return nil, err
	}
	return jsonData, nil
}

func (p *MyIdProvider) Endpoint() []byte {
	return []byte{}
}

func NewMyIdProvider(jwk *selfhosted.JWK, issuerHostPath string) *MyIdProvider {
	return &MyIdProvider{jwk, issuerHostPath}
}
