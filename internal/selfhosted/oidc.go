package selfhosted

import "context"

type OIDCIdPDiscoveryContents interface {
	Discovery() ([]byte, error)
	JWK() ([]byte, error)
	JWKsFileName() string
}

type OIDCIdPDiscovery interface {
	CreateStorage() error
	Upload(ctx context.Context, o OIDCIdPDiscoveryContents) error
	Endpoint() string
}

type OIDCIdP interface {
	Create(ctx context.Context) (string, error)
	IsUpdate() (bool, error)
	Update(ctx context.Context) error
}
