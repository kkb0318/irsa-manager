package client

import (
	"context"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials/stscreds"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/sts"
)

type AwsClient interface {
	IamClient() *AwsIamClient
	StsClient() *AwsStsClient
	S3Client(region, bucketName string) *AwsS3Client
}

type AwsIamClient struct {
	Client AwsIamAPI
}
type AwsStsClient struct {
	Client AwsStsAPI
}
type AwsS3Client struct {
	Client     AwsS3API
	region     string
	bucketName string
}

func NewAwsClientFactory(ctx context.Context) (*AwsClientFactory, error) {
	cfg, err := config.LoadDefaultConfig(
		ctx,
	)
	if err != nil {
		return nil, fmt.Errorf("unable to load SDK config, %w", err)
	}
	roleArn := os.Getenv("AWS_ROLE_ARN")
	if roleArn != "" {
		stsSvc := sts.NewFromConfig(cfg)
		creds := stscreds.NewAssumeRoleProvider(stsSvc, roleArn)
		cfg.Credentials = aws.NewCredentialsCache(creds)
	}
	return &AwsClientFactory{config: cfg}, nil
}

func (a *AwsClientFactory) IamClient() *AwsIamClient {
	return &AwsIamClient{
		iam.NewFromConfig(a.config),
	}
}

func (a *AwsClientFactory) StsClient() *AwsStsClient {
	return &AwsStsClient{
		sts.NewFromConfig(a.config),
	}
}

func (a *AwsClientFactory) S3Client(region, bucketName string) *AwsS3Client {
	return &AwsS3Client{
		Client:     s3.NewFromConfig(a.config),
		region:     region,
		bucketName: bucketName,
	}
}
