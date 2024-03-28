package oidc

type MyIdProvider struct {
}

func (p *MyIdProvider) Discovery() []byte {
  return []byte{}

}
func (p *MyIdProvider) JWK() []byte {
  return []byte{}

}
func (p *MyIdProvider) Endpoint() []byte {
  return []byte{}

}
