package selfhosted

import "context"

type OIDCIssuerMeta interface {
	IssuerHostPath() string
	IssuerUrl() string
}

type OIDCIdP interface {
	Create(ctx context.Context) error
	IsUpdate() (bool, error)
	Update(ctx context.Context) error
	Delete(ctx context.Context) error
}

type OIDCIdPDiscoveryContents interface {
	Discovery() ([]byte, error)
	JWK() ([]byte, error)
	JWKsFileName() string
}

type OIDCIdPDiscovery interface {
	CreateStorage(ctx context.Context) error
	Upload(ctx context.Context, o OIDCIdPDiscoveryContents, forceUpdate bool) error
	Delete(ctx context.Context, o OIDCIdPDiscoveryContents) error
}

type OIDCIdPFactory interface {
	IssuerMeta() OIDCIssuerMeta
	IdP(i OIDCIssuerMeta) (OIDCIdP, error)
	IdPDiscovery() OIDCIdPDiscovery
	IdPDiscoveryContents(i OIDCIssuerMeta) OIDCIdPDiscoveryContents
}
