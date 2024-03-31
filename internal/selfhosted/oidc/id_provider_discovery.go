package oidc

import (
	"context"
	"fmt"

	"github.com/kkb0318/irsa-manager/internal/client"
	"github.com/kkb0318/irsa-manager/internal/selfhosted"
)

const CONFIGURATION_PATH = ".well-known/openid-configuration"

type S3IdPDiscovery struct {
	s3Client *client.AwsS3Client
}

// NewS3IdPDiscovery initializes a new instance of S3IdPCreator with the specified AWS region and bucket name.
// This function attempts to create an AWS client configured for the specified region.
func NewS3IdPDiscovery(awsConfig *client.AwsConfig, bucketName string) (*S3IdPDiscovery, error) {
	s3Client := awsConfig.S3Cient(bucketName)
	return &S3IdPDiscovery{s3Client}, nil
}

// CreateStorage creates an S3 bucket
func (s *S3IdPDiscovery) CreateStorage() error {
	err := s.s3Client.CreateBucket(context.TODO())
	if err != nil {
		return fmt.Errorf("unable to create bucket, %w", err)
	}
	return nil
}

// Upload uploads the OIDC provider's discovery configuration and JSON Web Key Set (JWKS) to the specified AWS S3 bucket.
// This method is responsible for uploading the necessary OIDC configuration files to S3, making them accessible for OIDC clients.
func (s *S3IdPDiscovery) Upload(ctx context.Context, o selfhosted.OIDCIdPDiscoveryContents) error {
	discovery, err := o.Discovery()
	if err != nil {
		return nil
	}
	err = s.s3Client.PutObject(ctx,
		CONFIGURATION_PATH,
		discovery,
	)
	if err != nil {
		return fmt.Errorf("unable to upload discovery document, %w", err)
	}

	// Uplaod JWK
	jwk, err := o.JWK()
	if err != nil {
		return nil
	}
	err = s.s3Client.PutObject(ctx,
		o.JWKsFileName(),
		jwk,
	)
	if err != nil {
		return fmt.Errorf("unable to upload JWK, %w", err)
	}
	return nil
}

// Endpoint constructs the URL path for the OIDC issuer based on the provided AWS region and bucket name.
// This utility function generates the expected host path for accessing the OIDC configuration stored in an S3 bucket.
func (p *S3IdPDiscovery) Endpoint() string {
	return fmt.Sprintf("s3-%s.amazonaws.com/%s", p.s3Client.Region(), p.s3Client.BucketName())
}
