package selfhosted

import "context"

type OIDCIdProvider interface {
	Discovery() ([]byte, error)
	JWK() ([]byte, error)
	Endpoint() string
}

type OIDCIdPCreator interface {
	CreateStorage() error
	Upload(ctx context.Context, o OIDCIdProvider) error
}
