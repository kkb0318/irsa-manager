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

func NewS3IssuerMeta(s3 irsav1alpha1.S3Discovery) *S3IssuerMeta {
	return &S3IssuerMeta{s3.Region, s3.BucketName}
}

func (i *S3IssuerMeta) IssuerHostPath() string {
	return fmt.Sprintf("s3-%s.amazonaws.com/%s", i.region, i.bucketName)
}

// IssuerUrl constructs the URL path for the OIDC issuer based on the provided AWS region and bucket name.
// This utility function generates the expected host path for accessing the OIDC configuration stored in an S3 bucket.
func (i *S3IssuerMeta) IssuerUrl() string {
	return fmt.Sprintf("https://%s", i.
		IssuerHostPath())
}
