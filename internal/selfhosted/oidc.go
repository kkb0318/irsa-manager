package selfhosted

type OIDCIdProvider interface {
	Discovery() []byte
	JWK() []byte
	Endpoint() string
}

type OIDCIdPCreator interface {
	CreateStorage() error
	Upload(o OIDCIdProvider) error
}
