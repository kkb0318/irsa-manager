package oidc

import (
	"context"

	"github.com/kkb0318/irsa-manager/internal/client"
	"github.com/kkb0318/irsa-manager/internal/selfhosted"
)

type AwsIdP struct {
	iamClient *client.AwsIamClient
  discovery selfhosted.OIDCIdPDiscovery
}

func NewAwsIdP(awsConfig *client.AwsConfig, discovery selfhosted.OIDCIdPDiscovery) (*AwsIdP, error) {
	iamClient := awsConfig.IamCient()
	return &AwsIdP{iamClient, discovery}, nil
}

func (a *AwsIdP) Create(ctx context.Context) (string, error) {
	arn, err := a.iamClient.CreateOIDCProvider(ctx, a.discovery.Endpoint())
	if err != nil {
		return "", err
	}
	return arn, nil
}

func (a *AwsIdP) Update(ctx context.Context) error {
	return nil
}

func (a *AwsIdP) IsUpdate() (bool, error) {
	return false, nil
}
