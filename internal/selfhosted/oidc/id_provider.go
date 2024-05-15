package oidc

import (
	"context"

	awsclient "github.com/kkb0318/irsa-manager/internal/aws"
	"github.com/kkb0318/irsa-manager/internal/selfhosted"
)

type AwsIdP struct {
	iamClient  *awsclient.AwsIamClient
	stsClient  *awsclient.AwsStsClient
	issuerMeta selfhosted.OIDCIssuerMeta
}

func NewAwsIdP(awsConfig awsclient.AwsClient, issuerMeta selfhosted.OIDCIssuerMeta) (*AwsIdP, error) {
	iamClient := awsConfig.IamClient()
	stsClient := awsConfig.StsClient()
	return &AwsIdP{iamClient, stsClient, issuerMeta}, nil
}

func (a *AwsIdP) Create(ctx context.Context) error {
	err := a.iamClient.CreateOIDCProvider(ctx, a.issuerMeta.IssuerUrl())
	if err != nil {
		return err
	}
	return nil
}

func (a *AwsIdP) Update(ctx context.Context) error {
	return nil
}

func (a *AwsIdP) IsUpdate() (bool, error) {
	return false, nil
}

func (a *AwsIdP) Delete(ctx context.Context) error {
	accountId, err := a.stsClient.GetAccountId()
	if err != nil {
		return err
	}
	return a.iamClient.DeleteOIDCProvider(ctx, accountId, a.issuerMeta.IssuerHostPath())
}
