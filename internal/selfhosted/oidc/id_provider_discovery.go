package oidc

import (
	"context"
	"fmt"

	awsclient "github.com/kkb0318/irsa-manager/internal/aws"
	"github.com/kkb0318/irsa-manager/internal/selfhosted"
)

const CONFIGURATION_PATH = ".well-known/openid-configuration"

type S3IdPDiscovery struct {
	s3Client *awsclient.AwsS3Client
}

// NewS3IdPDiscovery initializes a new instance of S3IdPCreator with the specified AWS region and bucket name.
// This function attempts to create an AWS client configured for the specified region.
func NewS3IdPDiscovery(awsConfig awsclient.AwsClient, region, bucketName string) *S3IdPDiscovery {
	s3Client := awsConfig.S3Client(region, bucketName)
	return &S3IdPDiscovery{s3Client}
}

// CreateStorage creates an S3 bucket
func (s *S3IdPDiscovery) CreateStorage(ctx context.Context) error {
	err := s.s3Client.CreateBucketPublic(ctx)
	if err != nil {
		return fmt.Errorf("unable to create bucket, %w", err)
	}
	return nil
}

// Upload uploads the OIDC provider's discovery configuration and JSON Web Key Set (JWKS) to the specified AWS S3 bucket.
// This method is responsible for uploading the necessary OIDC configuration files to S3, making them accessible for OIDC clients.
func (s *S3IdPDiscovery) Upload(ctx context.Context, o selfhosted.OIDCIdPDiscoveryContents, forceUpdate bool) error {
	discovery, err := o.Discovery()
	if err != nil {
		return nil
	}
	jwk, err := o.JWK()
	if err != nil {
		return nil
	}
	inputs := []awsclient.ObjectInput{
		{
			Key:  CONFIGURATION_PATH,
			Body: discovery,
		},
		{
			Key:  o.JWKsFileName(),
			Body: jwk,
		},
	}
	if forceUpdate {
		err = s.s3Client.PutObjectsPublic(ctx, inputs)
	} else {
		err = s.s3Client.CreateObjectsPublic(ctx, inputs)
	}
	if err != nil {
		return fmt.Errorf("unable to upload object, %w", err)
	}
	return nil
}

// Delete delete an S3 bucket and objects
func (s *S3IdPDiscovery) Delete(ctx context.Context, o selfhosted.OIDCIdPDiscoveryContents) error {
	err := s.s3Client.DeleteObjects(ctx, []string{
		CONFIGURATION_PATH,
		o.JWKsFileName(),
	})
	if err != nil {
		return err
	}
	err = s.s3Client.DeleteBucket(ctx)
	if err != nil {
		return err
	}
	return nil
}
