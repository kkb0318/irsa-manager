package selfhosted

import "fmt"

type OIDCIdProvider interface {
	Discovery() string
	JWK() string
	Endpoint() string
}
type OIDCIdPCreator interface {
	Upload(o OIDCIdProvider) error
	CreateProvider() error
}

type S3IdPCreator struct {
	region     string
	bucketName string
}

func (s *S3IdPCreator) CreateProvider() error {
	// create S3 Bucket // bucketName, Region
	// create OIDCProvider //
	return nil
}

func (s *S3IdPCreator) Upload(o OIDCIdProvider) error {
	o.Discovery() // Upload to Endpoint()/.well-known/openid-configuration
	o.JWK()       // Upload to Endpoint()/keys.json
	return nil
}

func (s *S3IdPCreator) issuerHostPath() string {
	hostName := fmt.Sprintf("s3-%s.amazonaws.com", s.region)
	return fmt.Sprintf("%s/%s", hostName, s.bucketName)
}
