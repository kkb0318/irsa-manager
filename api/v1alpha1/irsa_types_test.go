package v1alpha1

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/types"
)

func TestStatusServiceAccountList_Append(t *testing.T) {
	tests := []struct {
		name     string
		initial  StatusServiceAccountList
		toAppend types.NamespacedName
		expected StatusServiceAccountList
	}{
		{
			name: "Append new item",
			initial: StatusServiceAccountList{
				{Name: "existing", Namespace: "default"},
			},
			toAppend: types.NamespacedName{Name: "new", Namespace: "default"},
			expected: StatusServiceAccountList{
				{Name: "existing", Namespace: "default"},
				{Name: "new", Namespace: "default"},
			},
		},
		{
			name: "Append existing item",
			initial: StatusServiceAccountList{
				{Name: "existing", Namespace: "default"},
			},
			toAppend: types.NamespacedName{Name: "existing", Namespace: "default"},
			expected: StatusServiceAccountList{
				{Name: "existing", Namespace: "default"},
				{Name: "existing", Namespace: "default"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.initial.Append(tt.toAppend)
			assert.Equal(t, tt.expected, tt.initial)
		})
	}
}

func TestStatusServiceAccountList_Delete(t *testing.T) {
	tests := []struct {
		name     string
		initial  StatusServiceAccountList
		toDelete types.NamespacedName
		expected StatusServiceAccountList
	}{
		{
			name: "Delete existing item",
			initial: StatusServiceAccountList{
				{Name: "existing", Namespace: "default"},
				{Name: "todelete", Namespace: "default"},
			},
			toDelete: types.NamespacedName{Name: "todelete", Namespace: "default"},
			expected: StatusServiceAccountList{
				{Name: "existing", Namespace: "default"},
			},
		},
		{
			name: "Delete non-existing item",
			initial: StatusServiceAccountList{
				{Name: "existing", Namespace: "default"},
			},
			toDelete: types.NamespacedName{Name: "nonexisting", Namespace: "default"},
			expected: StatusServiceAccountList{
				{Name: "existing", Namespace: "default"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.initial.Delete(tt.toDelete)
			assert.Equal(t, tt.expected, tt.initial)
		})
	}
}
