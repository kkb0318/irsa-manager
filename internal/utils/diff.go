package utils

import (
	"k8s.io/apimachinery/pkg/types"
)

// DiffNamespacedNames returns the namespaced names that are in target but not in reference.
func DiffNamespacedNames(target, reference []types.NamespacedName) []types.NamespacedName {
	referenceSet := make(map[types.NamespacedName]struct{})

	for _, item := range reference {
		referenceSet[item] = struct{}{}
	}

	diff := []types.NamespacedName{}
	for _, item := range target {
		if _, exists := referenceSet[item]; !exists {
			diff = append(diff, item)
		}
	}

	return diff
}
