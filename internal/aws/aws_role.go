package aws

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"slices"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/aws/smithy-go"
	irsav1alpha1 "github.com/kkb0318/irsa-manager/api/v1alpha1"
	"github.com/kkb0318/irsa-manager/internal/issuer"
)

// RoleManager represents the details needed to manage IAM roles
type RoleManager struct {
	// RoleName represents the name of the IAM role
	RoleName string
	// ServiceAccount represents the ServiceAccount Name and namespaces associated with the role
	ServiceAccount irsav1alpha1.IRSAServiceAccount
	// Policies represents the list of policies to be attached to the role
	Policies []string

	// AccountId represents the AWS Account Id
	AccountId string
}

// PolicyArn returns the full ARN of a given policy name. If the policy name already has the full ARN, it returns it as is.
func (r *RoleManager) PolicyArn(policy string) *string {
	prefix := "arn:aws:iam::"
	if strings.HasPrefix(policy, prefix) {
		return aws.String(policy)
	}
	return aws.String(fmt.Sprintf("%saws:policy/%s", prefix, policy))
}

// ExtractNewPolicies returns the names of the policies that are in the current settings (r.Policies) but are not yet attached to the role.
func (r *RoleManager) ExtractNewPolicies(l *iam.ListAttachedRolePoliciesOutput) []string {
	result := []string{}
	if l == nil {
		return r.Policies // return all policies
	}
	for _, p := range r.Policies {
		if !slices.ContainsFunc(l.AttachedPolicies, func(ap types.AttachedPolicy) bool {
			return *r.PolicyArn(p) == *ap.PolicyArn
		}) {
			result = append(result, p)
		}
	}
	return result
}

// ExtractStalePolicies returns the ARNs of the policies that are attached to the role but are not in the current settings (r.Policies).
func (r *RoleManager) ExtractStalePolicies(l *iam.ListAttachedRolePoliciesOutput) []string {
	result := []string{}
	if l == nil {
		return result
	}
	for _, ap := range l.AttachedPolicies {
		if !slices.ContainsFunc(r.Policies, func(p string) bool {
			return *r.PolicyArn(p) == *ap.PolicyArn
		}) {
			result = append(result, *ap.PolicyArn)
		}
	}
	return result
}

// DeleteIRSARole detaches specified policies from the IAM role and deletes the IAM role
func (a *AwsIamClient) DeleteIRSARole(ctx context.Context, r RoleManager) error {
	for _, policy := range r.Policies {
		err := a.DetachRolePolicy(ctx, aws.String(r.RoleName), r.PolicyArn(policy))
		if err != nil {
			return err
		}
		log.Printf("Policy %s detached from role %s successfully", policy, r.RoleName)
	}
	input := &iam.DeleteRoleInput{RoleName: aws.String(r.RoleName)}
	_, err := a.Client.DeleteRole(ctx, input)
	// Ignore error if the role does not exist or there are other policies that this controller does not manage
	if errorHandler(err, []string{"DeleteConflict", "NoSuchEntity"}) != nil {
		return err
	}
	log.Printf("Role %s deleted successfully", r.RoleName)
	return nil
}

// DetachRolePolicy detaches specified policies from the IAM role
func (a *AwsIamClient) DetachRolePolicy(ctx context.Context, roleName, policyArn *string) error {
	detachRolePolicyInput := &iam.DetachRolePolicyInput{
		RoleName:  roleName,
		PolicyArn: policyArn,
	}
	_, err := a.Client.DetachRolePolicy(ctx, detachRolePolicyInput)
	// Ignore error if the policy is already detached or the role does not exist
	if errorHandler(err, []string{"NoSuchEntity"}) != nil {
		return err
	}
	return nil
}

// AttachRolePolicy attaches specidied policy
func (a *AwsIamClient) AttachRolePolicy(ctx context.Context, roleName, policyArn *string) error {
	attachRolePolicyInput := &iam.AttachRolePolicyInput{
		RoleName:  roleName,
		PolicyArn: policyArn,
	}
	_, err := a.Client.AttachRolePolicy(ctx, attachRolePolicyInput)
	return err
}

// UpdateIRSARole creates an IAM role with the specified trust policy and attaches specified policies to it
func (a *AwsIamClient) UpdateIRSARole(ctx context.Context, issuerMeta issuer.OIDCIssuerMeta, r RoleManager) error {
	providerArn := fmt.Sprintf("arn:aws:iam::%s:oidc-provider/%s", r.AccountId, issuerMeta.IssuerHostPath())
	statement := make([]map[string]interface{}, len(r.ServiceAccount.Namespaces))
	for i, ns := range r.ServiceAccount.Namespaces {
		statement[i] = map[string]interface{}{
			"Effect": "Allow",
			"Principal": map[string]interface{}{
				"Federated": providerArn,
			},
			"Action": "sts:AssumeRoleWithWebIdentity",
			"Condition": map[string]interface{}{
				"StringEquals": map[string]interface{}{
					fmt.Sprintf("%s:sub", issuerMeta.IssuerHostPath()): fmt.Sprintf("system:serviceaccount:%s:%s", ns, r.ServiceAccount.Name),
				},
			},
		}
	}
	trustPolicy := map[string]interface{}{
		"Version":   "2012-10-17",
		"Statement": statement,
	}
	trustPolicyJSON, err := json.Marshal(trustPolicy)
	if err != nil {
		return fmt.Errorf("failed to marshal trust policy: %w", err)
	}
	createRoleInput := &iam.CreateRoleInput{
		RoleName:                 aws.String(r.RoleName),
		AssumeRolePolicyDocument: aws.String(string(trustPolicyJSON)),
	}

	_, err = a.Client.CreateRole(ctx, createRoleInput)
	if errorHandler(err, []string{"EntityAlreadyExists"}) != nil {
		return err
	}
	log.Printf("Role %s created successfully", r.RoleName)

	updateRoleInput := &iam.UpdateAssumeRolePolicyInput{
		RoleName:       aws.String(r.RoleName),
		PolicyDocument: aws.String(string(trustPolicyJSON)),
	}

	_, err = a.Client.UpdateAssumeRolePolicy(ctx, updateRoleInput)
	if err != nil {
		return fmt.Errorf("failed to update assume role policy for role %s: %w", r.RoleName, err)
	}

	listPoliciesOutput, err := a.Client.ListAttachedRolePolicies(ctx, &iam.ListAttachedRolePoliciesInput{RoleName: aws.String(r.RoleName)})
	if err != nil {
		return fmt.Errorf("failed to list attached role policies with %s: %w", r.RoleName, err)
	}

	for _, policy := range r.ExtractNewPolicies(listPoliciesOutput) {
		err := a.AttachRolePolicy(ctx, aws.String(r.RoleName), r.PolicyArn(policy))
		if err != nil {
			return err
		}
		log.Printf("Policy %s attached to role %s successfully", policy, r.RoleName)
	}
	for _, policy := range r.ExtractStalePolicies(listPoliciesOutput) {
		err := a.DetachRolePolicy(ctx, aws.String(r.RoleName), r.PolicyArn(policy))
		if err != nil {
			return err
		}
		log.Printf("Policy %s detached to role %s successfully", policy, r.RoleName)
	}
	log.Printf("Assume role policy for %s updated successfully", r.RoleName)
	return nil
}

// errorHandler handles specific errors by checking the error code against a list of codes to ignore
func errorHandler(err error, errorCodes []string) error {
	if err != nil {
		var ae smithy.APIError
		if errors.As(err, &ae) && slices.Contains(errorCodes, ae.ErrorCode()) {
			// fmt.Printf("Skipped error: %s \n", err.Error())
			return nil
		}
	}
	return err
}
