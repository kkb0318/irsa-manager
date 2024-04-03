package client

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

type AwsConfig struct {
	config aws.Config
}

func NewAwsClient(ctx context.Context, region string) (*AwsConfig, error) {
	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion(region),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to load SDK config, %w", err)
	}
	return &AwsConfig{config: cfg}, nil
}

func (a *AwsConfig) IamCient() *AwsIamClient {
	return &AwsIamClient{
		iam.NewFromConfig(a.config),
	}
}

func (a *AwsConfig) S3Cient(bucketName string) *AwsS3Client {
	return &AwsS3Client{
		bucketName,
		s3.NewFromConfig(a.config),
	}
}

type AwsS3Client struct {
	bucketName string
	client     *s3.Client
}

func (a *AwsS3Client) PutObject(ctx context.Context, key string, body []byte) error {
	_, err := a.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(a.bucketName),
		Key:    aws.String(key),
		Body:   bytes.NewReader(body),
	})
	return err
}

func (a *AwsS3Client) CreateBucket(ctx context.Context) error {
	_, err := a.client.CreateBucket(ctx, &s3.CreateBucketInput{
		ACL:    types.BucketCannedACLPublicRead,
		Bucket: aws.String(a.bucketName),
		CreateBucketConfiguration: &types.CreateBucketConfiguration{
			LocationConstraint: types.BucketLocationConstraint(a.Region()),
		},
		ObjectOwnership: types.ObjectOwnershipBucketOwnerPreferred,
	})
	return err
}

func (a *AwsS3Client) BucketName() string {
	return a.bucketName
}

func (a *AwsS3Client) Region() string {
	return a.client.Options().Region
}

type AwsIamClient struct {
	client *iam.Client
}

func (a *AwsIamClient) CreateOIDCProvider(ctx context.Context, providerUrl string) (string, error) {
	result, err := a.client.CreateOpenIDConnectProvider(ctx, &iam.CreateOpenIDConnectProviderInput{
		Url:          &providerUrl,
		ClientIDList: []string{"sts.amazonaws.com"},
		ThumbprintList: []string{
			strings.Repeat("x", 40), // Thumbprint is required, but IAM will retrieve and use the top intermediate CA thumbprint of the OpenID Connect identity provider server certificate.
		},
	})
	if err != nil {
		return "", err
	}
	return *result.OpenIDConnectProviderArn, nil
}

func (a *AwsIamClient) Region() string {
	return a.client.Options().Region
}
