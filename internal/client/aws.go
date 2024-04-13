package client

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/smithy-go"
)

type AwsClientFactory struct {
	config aws.Config
}

type AwsIamAPI interface {
	CreateOpenIDConnectProvider(ctx context.Context, params *iam.CreateOpenIDConnectProviderInput, optFns ...func(*iam.Options)) (*iam.CreateOpenIDConnectProviderOutput, error)
}

type AwsS3API interface {
	CreateBucket(ctx context.Context, params *s3.CreateBucketInput, optFns ...func(*s3.Options)) (*s3.CreateBucketOutput, error)
	PutObject(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error)
	DeletePublicAccessBlock(ctx context.Context, params *s3.DeletePublicAccessBlockInput, optFns ...func(*s3.Options)) (*s3.DeletePublicAccessBlockOutput, error)
	PutBucketOwnershipControls(ctx context.Context, params *s3.PutBucketOwnershipControlsInput, optFns ...func(*s3.Options)) (*s3.PutBucketOwnershipControlsOutput, error)
	DeleteBucket(ctx context.Context, params *s3.DeleteBucketInput, optFns ...func(*s3.Options)) (*s3.DeleteBucketOutput, error)
	DeleteObjects(ctx context.Context, params *s3.DeleteObjectsInput, optFns ...func(*s3.Options)) (*s3.DeleteObjectsOutput, error)
}

type AwsClient interface {
	IamClient() *AwsIamClient
	S3Client(region, bucketName string) *AwsS3Client
}

func NewAwsClientFactory(ctx context.Context) (*AwsClientFactory, error) {
	cfg, err := config.LoadDefaultConfig(
		ctx,
	)
	if err != nil {
		return nil, fmt.Errorf("unable to load SDK config, %w", err)
	}
	return &AwsClientFactory{config: cfg}, nil
}

func (a *AwsClientFactory) IamClient() *AwsIamClient {
	return &AwsIamClient{
		iam.NewFromConfig(a.config),
	}
}

func (a *AwsClientFactory) S3Client(bucketName, region string) *AwsS3Client {
	return &AwsS3Client{
		s3.NewFromConfig(a.config),
		region,
		bucketName,
	}
}

type AwsS3Client struct {
	Client     AwsS3API
	region     string
	bucketName string
}

// PutObjectPublic uploads a file to an S3 bucket and sets its access level to public read.
// This means the file can be read by anyone on the internet.
func (a *AwsS3Client) PutObjectPublic(ctx context.Context, key string, body []byte) error {
	_, err := a.Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(a.bucketName),
		Key:         aws.String(key),
		ACL:         types.ObjectCannedACLPublicRead,
		Body:        bytes.NewReader(body),
		ContentType: aws.String("application/json"),
	})
	return err
}

// CreateBucketPublic creates a new S3 bucket with public access settings in the specified region.
// The function configures the bucket to have its ownership controlled by the bucket creator.
func (a *AwsS3Client) CreateBucketPublic(ctx context.Context) error {
	bucket := aws.String(a.bucketName)
	_, err := a.Client.CreateBucket(ctx, &s3.CreateBucketInput{
		Bucket: bucket,
		CreateBucketConfiguration: &types.CreateBucketConfiguration{
			LocationConstraint: types.BucketLocationConstraint(a.Region()),
		},
	})
	if err != nil {
		var bucketAlreadyOwnedByYou *types.BucketAlreadyOwnedByYou
		if errors.As(err, &bucketAlreadyOwnedByYou) {
			log.Println("skipped error", err)
		} else {
			return err
		}
	}
	_, err = a.Client.DeletePublicAccessBlock(ctx, &s3.DeletePublicAccessBlockInput{Bucket: bucket})
	if err != nil {
		return err
	}
	_, err = a.Client.PutBucketOwnershipControls(ctx, &s3.PutBucketOwnershipControlsInput{
		Bucket: bucket,
		OwnershipControls: &types.OwnershipControls{
			Rules: []types.OwnershipControlsRule{
				{
					ObjectOwnership: types.ObjectOwnershipBucketOwnerPreferred,
				},
			},
		},
	})
	if err != nil {
		return err
	}
	return nil
}

// DeleteBucket attempts to delete the specified bucket.
// If the bucket contains any objects, the deletion will not be forced to prevent accidental data loss.
func (a *AwsS3Client) DeleteBucket(ctx context.Context) error {
	_, err := a.Client.DeleteBucket(ctx, &s3.DeleteBucketInput{
		Bucket: aws.String(a.bucketName),
	})
	if err != nil {
		var ae smithy.APIError
		if errors.As(err, &ae) && ae.ErrorCode() == "BucketNotEmpty" {
			log.Println("skipped error", err)
		} else {
			return err
		}
	}
	return nil
}

// DeleteObjects removes a list of objects from a specified bucket.
func (a *AwsS3Client) DeleteObjects(ctx context.Context, objectKeys []string) error {
	var objectIds []types.ObjectIdentifier
	for _, key := range objectKeys {
		objectIds = append(objectIds, types.ObjectIdentifier{Key: aws.String(key)})
	}
	_, err := a.Client.DeleteObjects(ctx, &s3.DeleteObjectsInput{
		Bucket: aws.String(a.bucketName),
		Delete: &types.Delete{Objects: objectIds},
	})
	if err != nil {
		return err
	}
	return err
}

func (a *AwsS3Client) BucketName() string {
	return a.bucketName
}

func (a *AwsS3Client) Region() string {
	return a.region
}

type AwsIamClient struct {
	Client AwsIamAPI
}

// CreateOIDCProvider creates an OpenID Connect (OIDC) provider in AWS IAM.
func (a *AwsIamClient) CreateOIDCProvider(ctx context.Context, providerUrl string) (string, error) {
	result, err := a.Client.CreateOpenIDConnectProvider(ctx, &iam.CreateOpenIDConnectProviderInput{
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
