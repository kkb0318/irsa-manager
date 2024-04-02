package oidc

import (
	"context"

	awsclient "github.com/kkb0318/irsa-manager/internal/client"
	"github.com/kkb0318/irsa-manager/internal/selfhosted"
)

type AwsS3IdPFactory struct {
	region       string
	bucketName   string
	awsConfig    *awsclient.AwsConfig
	jwk          *selfhosted.JWK
	jwksFileName string
}

func NewAwsS3IdpFactory(ctx context.Context, region, bucketName string, jwk *selfhosted.JWK, jwksFileName string) (*AwsS3IdPFactory, error) {
	awsConfig, err := awsclient.NewAwsClient(ctx, region)
	if err != nil {
		return nil, err
	}
	return &AwsS3IdPFactory{
		region,
		bucketName,
		awsConfig,
		jwk,
		jwksFileName,
	}, nil
}

func (f *AwsS3IdPFactory) IssuerMeta() selfhosted.OIDCIssuerMeta {
	return NewS3IssuerMeta(f.region, f.bucketName)
}

func (f *AwsS3IdPFactory) IdP(i selfhosted.OIDCIssuerMeta) (selfhosted.OIDCIdP, error) {
	return NewAwsIdP(f.awsConfig, i)
}

func (f *AwsS3IdPFactory) IdPDiscovery() selfhosted.OIDCIdPDiscovery {
	return NewS3IdPDiscovery(f.awsConfig, f.bucketName)
}

func (f *AwsS3IdPFactory) IdPDiscoveryContents(i selfhosted.OIDCIssuerMeta) selfhosted.OIDCIdPDiscoveryContents {
	return NewIdPDiscoveryContents(f.jwk, i, f.jwksFileName)
}
