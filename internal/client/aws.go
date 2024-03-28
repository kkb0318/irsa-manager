package client

import (
	"bytes"
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
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
		Bucket: aws.String(a.bucketName),
	})
	return err
}
