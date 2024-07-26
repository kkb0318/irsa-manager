package issuer

import (
	"fmt"

	irsav1alpha1 "github.com/kkb0318/irsa-manager/api/v1alpha1"
)

type OIDCIssuerMeta interface {
	IssuerHostPath() string
	IssuerUrl() string
}

type S3IssuerMeta struct {
	region     string
	bucketName string
}

func NewOIDCIssuerMeta(i *irsav1alpha1.IRSASetup) (OIDCIssuerMeta, error) {
	if i.Spec.Mode == irsav1alpha1.ModeEks {
		return newIamOIDCProviderIssuerMeta(i.Spec.IamOIDCProvider)
	}
	return NewS3IssuerMeta(&i.Spec.Discovery.S3)
}

func NewS3IssuerMeta(s3 *irsav1alpha1.S3Discovery) (*S3IssuerMeta, error) {
	region := s3.Region
	bucketName := s3.BucketName
	if region == "" || bucketName == "" {
		return nil, fmt.Errorf("s3 region and bucket name must not be empty. region: %s, bucketName: %s", region, bucketName)
	}
	return &S3IssuerMeta{region, bucketName}, nil
}

func (i *S3IssuerMeta) IssuerHostPath() string {
	return fmt.Sprintf("s3-%s.amazonaws.com/%s", i.region, i.bucketName)
}

// IssuerUrl constructs the URL path for the OIDC issuer based on the provided AWS region and bucket name.
// This utility function generates the expected host path for accessing the OIDC configuration stored in an S3 bucket.
func (i *S3IssuerMeta) IssuerUrl() string {
	return fmt.Sprintf("https://%s", i.IssuerHostPath())
}

func newIamOIDCProviderIssuerMeta(providerName string) (*iamOIDCProviderIssuerMeta, error) {
	if providerName == "" {
		return nil, fmt.Errorf("IAM OIDC Provider Name must not be empty")
	}
	return &iamOIDCProviderIssuerMeta{providerName}, nil
}

type iamOIDCProviderIssuerMeta struct {
	providerName string
}

func (i *iamOIDCProviderIssuerMeta) IssuerHostPath() string {
	return i.providerName
}

func (i *iamOIDCProviderIssuerMeta) IssuerUrl() string {
	return fmt.Sprintf("https://%s", i.IssuerHostPath())
}
