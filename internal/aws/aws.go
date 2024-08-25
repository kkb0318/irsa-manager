package aws

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log"
	"slices"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	iamtypes "github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/aws/smithy-go"
)

type AwsClientFactory struct {
	config aws.Config
}

type AwsIamAPI interface {
	CreateOpenIDConnectProvider(ctx context.Context, params *iam.CreateOpenIDConnectProviderInput, optFns ...func(*iam.Options)) (*iam.CreateOpenIDConnectProviderOutput, error)
	DeleteOpenIDConnectProvider(ctx context.Context, params *iam.DeleteOpenIDConnectProviderInput, optFns ...func(*iam.Options)) (*iam.DeleteOpenIDConnectProviderOutput, error)
	CreateRole(ctx context.Context, params *iam.CreateRoleInput, optFns ...func(*iam.Options)) (*iam.CreateRoleOutput, error)
	UpdateAssumeRolePolicy(ctx context.Context, params *iam.UpdateAssumeRolePolicyInput, optFns ...func(*iam.Options)) (*iam.UpdateAssumeRolePolicyOutput, error)
	ListAttachedRolePolicies(ctx context.Context, params *iam.ListAttachedRolePoliciesInput, optFns ...func(*iam.Options)) (*iam.ListAttachedRolePoliciesOutput, error)
	AttachRolePolicy(ctx context.Context, params *iam.AttachRolePolicyInput, optFns ...func(*iam.Options)) (*iam.AttachRolePolicyOutput, error)
	DeleteRole(ctx context.Context, params *iam.DeleteRoleInput, optFns ...func(*iam.Options)) (*iam.DeleteRoleOutput, error)
	DetachRolePolicy(ctx context.Context, params *iam.DetachRolePolicyInput, optFns ...func(*iam.Options)) (*iam.DetachRolePolicyOutput, error)
}

type AwsStsAPI interface {
	GetCallerIdentity(ctx context.Context, params *sts.GetCallerIdentityInput, optFns ...func(*sts.Options)) (*sts.GetCallerIdentityOutput, error)
}

type AwsS3API interface {
	CreateBucket(ctx context.Context, params *s3.CreateBucketInput, optFns ...func(*s3.Options)) (*s3.CreateBucketOutput, error)
	DeletePublicAccessBlock(ctx context.Context, params *s3.DeletePublicAccessBlockInput, optFns ...func(*s3.Options)) (*s3.DeletePublicAccessBlockOutput, error)
	DeleteBucket(ctx context.Context, params *s3.DeleteBucketInput, optFns ...func(*s3.Options)) (*s3.DeleteBucketOutput, error)
	DeleteObjects(ctx context.Context, params *s3.DeleteObjectsInput, optFns ...func(*s3.Options)) (*s3.DeleteObjectsOutput, error)
	HeadObject(ctx context.Context, params *s3.HeadObjectInput, optFns ...func(*s3.Options)) (*s3.HeadObjectOutput, error)
	PutObject(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error)
	PutBucketOwnershipControls(ctx context.Context, params *s3.PutBucketOwnershipControlsInput, optFns ...func(*s3.Options)) (*s3.PutBucketOwnershipControlsOutput, error)
}

// CheckObjectExists checks if a specific object exists in the given bucket.
func (a *AwsS3Client) CheckObjectExists(ctx context.Context, key string) (bool, error) {
	_, err := a.Client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(a.bucketName),
		Key:    aws.String(key),
	})
	if err != nil {
		var nfe *s3types.NotFound
		if errors.As(err, &nfe) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

type ObjectInput struct {
	Key  string
	Body []byte
}

func (a *AwsS3Client) CreateObjectsPublic(ctx context.Context, inputs []ObjectInput) error {
	for _, input := range inputs {
		if err := a.CreateObjectPublic(ctx, input); err != nil {
			return err
		}
	}
	return nil
}

// CreateObjectPublic creates a file to an S3 bucket and sets its access level to public read.
// This means the file can be read by anyone on the internet.
func (a *AwsS3Client) CreateObjectPublic(ctx context.Context, input ObjectInput) error {
	exists, err := a.CheckObjectExists(ctx, input.Key)
	if err != nil {
		return err
	}
	if exists {
		log.Printf("skipped to create bucket object %s \n", input.Key)
	} else {
		err := a.PutObjectPublic(ctx, input)
		if err != nil {
			return err
		}
	}
	return nil
}

func (a *AwsS3Client) PutObjectsPublic(ctx context.Context, inputs []ObjectInput) error {
	for _, input := range inputs {
		if err := a.PutObjectPublic(ctx, input); err != nil {
			return err
		}
	}
	return nil
}

// PutObjectPublic uploads a file to an S3 bucket and sets its access level to public read.
// This means the file can be read by anyone on the internet.
func (a *AwsS3Client) PutObjectPublic(ctx context.Context, input ObjectInput) error {
	_, err := a.Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(a.bucketName),
		Key:         aws.String(input.Key),
		ACL:         s3types.ObjectCannedACLPublicRead,
		Body:        bytes.NewReader(input.Body),
		ContentType: aws.String("application/json"),
	})
	return err
}

// CreateBucketPublic creates a new S3 bucket with public access settings in the specified region.
// The function configures the bucket to have its ownership controlled by the bucket creator.
func (a *AwsS3Client) CreateBucketPublic(ctx context.Context) error {
	log.Printf("creating S3 bucket... Name: %s, Region: %s \n", a.bucketName, a.Region())
	bucket := aws.String(a.bucketName)
	var input *s3.CreateBucketInput
	if a.Region() == "us-east-1" {
		input = &s3.CreateBucketInput{
			Bucket: bucket,
		}
	} else {
		input = &s3.CreateBucketInput{
			Bucket: bucket,
			CreateBucketConfiguration: &s3types.CreateBucketConfiguration{
				LocationConstraint: s3types.BucketLocationConstraint(a.Region()),
			},
		}
	}
	_, err := a.Client.CreateBucket(ctx, input)
	if err != nil {
		var bucketAlreadyOwnedByYou *s3types.BucketAlreadyOwnedByYou
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
		OwnershipControls: &s3types.OwnershipControls{
			Rules: []s3types.OwnershipControlsRule{
				{
					ObjectOwnership: s3types.ObjectOwnershipBucketOwnerPreferred,
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
		errorCodes := []string{"BucketNotEmpty", "NoSuchBucket"}
		if errors.As(err, &ae) && slices.Contains(errorCodes, ae.ErrorCode()) {
			log.Println("Deletion skipped: ", err)
			return nil
		}
		return err
	}
	return nil
}

// DeleteObjects removes a list of objects from a specified bucket.
func (a *AwsS3Client) DeleteObjects(ctx context.Context, objectKeys []string) error {
	objectIds := make([]s3types.ObjectIdentifier, len(objectKeys))
	for i, key := range objectKeys {
		objectIds[i] = s3types.ObjectIdentifier{Key: aws.String(key)}
	}
	_, err := a.Client.DeleteObjects(ctx, &s3.DeleteObjectsInput{
		Bucket: aws.String(a.bucketName),
		Delete: &s3types.Delete{Objects: objectIds},
	})
	if err != nil {
		var ae smithy.APIError
		if errors.As(err, &ae) && ae.ErrorCode() == "NoSuchBucket" {
			log.Println("Deletion skipped: the bucket does not exist.", err)
			return nil
		}
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

// CreateOIDCProvider creates an OpenID Connect (OIDC) provider in AWS IAM.
func (a *AwsIamClient) CreateOIDCProvider(ctx context.Context, providerUrl string) error {
	_, err := a.Client.CreateOpenIDConnectProvider(ctx, &iam.CreateOpenIDConnectProviderInput{
		Url:          &providerUrl,
		ClientIDList: []string{"sts.amazonaws.com"},
		ThumbprintList: []string{
			strings.Repeat("x", 40), // Thumbprint is required, but IAM will retrieve and use the top intermediate CA thumbprint of the OpenID Connect identity provider server certificate.
		},
	})
	if err != nil {
		var entityAlreadyExists *iamtypes.EntityAlreadyExistsException
		if errors.As(err, &entityAlreadyExists) {
			log.Println("skipped error", err)
		} else {
			return err
		}
	}
	return nil
}

// DeleteOIDCProvider deletes an OpenID Connect (OIDC) provider in AWS IAM.
func (a *AwsIamClient) DeleteOIDCProvider(ctx context.Context, accountId, issuerHostPath string) error {
	_, err := a.Client.DeleteOpenIDConnectProvider(ctx, &iam.DeleteOpenIDConnectProviderInput{
		OpenIDConnectProviderArn: aws.String(fmt.Sprintf("arn:aws:iam::%s:oidc-provider/%s", accountId, issuerHostPath)),
	})
	if err != nil {
		var ae smithy.APIError
		if errors.As(err, &ae) && ae.ErrorCode() == "NoSuchEntity" {
			log.Println("Deletion skipped: ", err)
			return nil
		}
		return err
	}
	return nil
}

func (a *AwsStsClient) GetAccountId() (string, error) {
	req, err := a.Client.GetCallerIdentity(context.TODO(), &sts.GetCallerIdentityInput{})
	if err != nil {
		return "", err
	}

	return *req.Account, nil
}
