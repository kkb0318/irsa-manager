package selfhosted

type OIDCIdProvider interface {
	Discovery() ([]byte, error)
	JWK() ([]byte, error)
	Endpoint() string
}

type OIDCIdPCreator interface {
	CreateStorage() error
	Upload(o OIDCIdProvider) error
}
