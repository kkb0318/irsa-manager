package selfhosted

import "context"

func Execute(ctx context.Context, idpComponentsFactory OIDCIdPFactory, forceUpdate bool) error {
	issuerMeta := idpComponentsFactory.IssuerMeta()
	discovery := idpComponentsFactory.IdPDiscovery()
	discoveryContents := idpComponentsFactory.IdPDiscoveryContents(issuerMeta)
	idp, err := idpComponentsFactory.IdP(issuerMeta)
	if err != nil {
		return err
	}
	err = discovery.CreateStorage(ctx)
	if err != nil {
		return err
	}
	err = discovery.Upload(ctx, discoveryContents, forceUpdate)
	if err != nil {
		return err
	}
	err = idp.Create(ctx)
	if err != nil {
		return err
	}
	return nil
}

func Delete(ctx context.Context, factory OIDCIdPFactory) error {
	issuerMeta := factory.IssuerMeta()
	discovery := factory.IdPDiscovery()
	discoveryContents := factory.IdPDiscoveryContents(issuerMeta)
	idp, err := factory.IdP(issuerMeta)
	if err != nil {
		return err
	}
	err = discovery.Delete(ctx, discoveryContents)
	if err != nil {
		return err
	}
	err = idp.Delete(ctx)
	if err != nil {
		return err
	}
	return nil
}
