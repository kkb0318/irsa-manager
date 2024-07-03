package aws

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/stretchr/testify/assert"
)

func TestExtractNewPolicies(t *testing.T) {
	tests := []struct {
		name             string
		policies         []string
		attachedPolicies *iam.ListAttachedRolePoliciesOutput
		expected         []string
	}{
		{
			"PolicyAlreadyAttached",
			[]string{"ReadOnlyAccess", "AdministratorAccess"},
			&iam.ListAttachedRolePoliciesOutput{
				AttachedPolicies: []types.AttachedPolicy{
					{PolicyArn: aws.String("arn:aws:iam::aws:policy/ReadOnlyAccess")},
				},
			},
			[]string{"AdministratorAccess"},
		},
		{
			"NoPolicyAttached",
			[]string{"PowerUserAccess", "SecurityAudit"},
			&iam.ListAttachedRolePoliciesOutput{
				AttachedPolicies: []types.AttachedPolicy{
					{PolicyArn: aws.String("arn:aws:iam::aws:policy/ReadOnlyAccess")},
				},
			},
			[]string{"PowerUserAccess", "SecurityAudit"},
		},
		{
			"AllPoliciesAlreadyAttached",
			[]string{"ReadOnlyAccess"},
			&iam.ListAttachedRolePoliciesOutput{
				AttachedPolicies: []types.AttachedPolicy{
					{PolicyArn: aws.String("arn:aws:iam::aws:policy/ReadOnlyAccess")},
				},
			},
			[]string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &RoleManager{Policies: tt.policies}
			result := r.ExtractNewPolicies(tt.attachedPolicies)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestExtractStalePolicies(t *testing.T) {
	tests := []struct {
		name             string
		policies         []string
		attachedPolicies *iam.ListAttachedRolePoliciesOutput
		expected         []string
	}{
		{
			"StalePolicyExists",
			[]string{"ReadOnlyAccess"},
			&iam.ListAttachedRolePoliciesOutput{
				AttachedPolicies: []types.AttachedPolicy{
					{PolicyArn: aws.String("arn:aws:iam::aws:policy/ReadOnlyAccess")},
					{PolicyArn: aws.String("arn:aws:iam::aws:policy/AdministratorAccess")},
				},
			},
			[]string{"arn:aws:iam::aws:policy/AdministratorAccess"},
		},
		{
			"MultipleStalePoliciesExist",
			[]string{"ReadOnlyAccess", "SecurityAudit"},
			&iam.ListAttachedRolePoliciesOutput{
				AttachedPolicies: []types.AttachedPolicy{
					{PolicyArn: aws.String("arn:aws:iam::aws:policy/PowerUserAccess")},
					{PolicyArn: aws.String("arn:aws:iam::aws:policy/AdministratorAccess")},
				},
			},
			[]string{"arn:aws:iam::aws:policy/PowerUserAccess", "arn:aws:iam::aws:policy/AdministratorAccess"},
		},
		{
			"NoStalePolicies",
			[]string{"ReadOnlyAccess"},
			&iam.ListAttachedRolePoliciesOutput{
				AttachedPolicies: []types.AttachedPolicy{
					{PolicyArn: aws.String("arn:aws:iam::aws:policy/ReadOnlyAccess")},
				},
			},
			[]string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &RoleManager{Policies: tt.policies}
			result := r.ExtractStalePolicies(tt.attachedPolicies)
			assert.Equal(t, tt.expected, result)
		})
	}
}
