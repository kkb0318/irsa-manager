package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/types"
)

func TestDiffNamespacedNames(t *testing.T) {
	tests := []struct {
		name      string
		target    []types.NamespacedName
		reference []types.NamespacedName
		expected  []types.NamespacedName
	}{
		{
			"NoDifference",
			[]types.NamespacedName{
				{Namespace: "default", Name: "resource1"},
				{Namespace: "default", Name: "resource2"},
			},
			[]types.NamespacedName{
				{Namespace: "default", Name: "resource1"},
				{Namespace: "default", Name: "resource2"},
			},
			[]types.NamespacedName{},
		},
		{
			"SomeDifference",
			[]types.NamespacedName{
				{Namespace: "default", Name: "resource1"},
				{Namespace: "default", Name: "resource3"},
			},
			[]types.NamespacedName{
				{Namespace: "default", Name: "resource1"},
				{Namespace: "default", Name: "resource2"},
			},
			[]types.NamespacedName{
				{Namespace: "default", Name: "resource3"},
			},
		},
		{
			"AllDifferent",
			[]types.NamespacedName{
				{Namespace: "default", Name: "resource1"},
				{Namespace: "default", Name: "resource2"},
			},
			[]types.NamespacedName{
				{Namespace: "other", Name: "resource3"},
				{Namespace: "other", Name: "resource4"},
			},
			[]types.NamespacedName{
				{Namespace: "default", Name: "resource1"},
				{Namespace: "default", Name: "resource2"},
			},
		},
		{
			"EmptyTarget",
			[]types.NamespacedName{},
			[]types.NamespacedName{
				{Namespace: "default", Name: "resource1"},
				{Namespace: "default", Name: "resource2"},
			},
			[]types.NamespacedName{},
		},
		{
			"EmptyReference",
			[]types.NamespacedName{
				{Namespace: "default", Name: "resource1"},
				{Namespace: "default", Name: "resource2"},
			},
			[]types.NamespacedName{},
			[]types.NamespacedName{
				{Namespace: "default", Name: "resource1"},
				{Namespace: "default", Name: "resource2"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DiffNamespacedNames(tt.target, tt.reference)
			assert.Equal(t, tt.expected, result)
		})
	}
}
