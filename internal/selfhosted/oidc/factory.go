package oidc

import (
	"context"

	awsclient "github.com/kkb0318/irsa-manager/internal/aws"
	"github.com/kkb0318/irsa-manager/internal/issuer"
	"github.com/kkb0318/irsa-manager/internal/selfhosted"
)

type AwsS3IdPFactory struct {
	region       string
	bucketName   string
	awsClient    awsclient.AwsClient
	jwk          *selfhosted.JWK
	jwksFileName string
}

func NewAwsS3IdpFactory(
	ctx context.Context,
	region,
	bucketName string,
	jwk *selfhosted.JWK,
	jwksFileName string,
	awsClient awsclient.AwsClient,
) (*AwsS3IdPFactory, error) {
	return &AwsS3IdPFactory{
		region:       region,
		bucketName:   bucketName,
		awsClient:    awsClient,
		jwk:          jwk,
		jwksFileName: jwksFileName,
	}, nil
}

func (f *AwsS3IdPFactory) IdP(i issuer.OIDCIssuerMeta) (selfhosted.OIDCIdP, error) {
	return NewAwsIdP(f.awsClient, i)
}

func (f *AwsS3IdPFactory) IdPDiscovery() selfhosted.OIDCIdPDiscovery {
	return NewS3IdPDiscovery(f.awsClient, f.region, f.bucketName)
}

func (f *AwsS3IdPFactory) IdPDiscoveryContents(i issuer.OIDCIssuerMeta) selfhosted.OIDCIdPDiscoveryContents {
	return NewIdPDiscoveryContents(f.jwk, i, f.jwksFileName)
}
