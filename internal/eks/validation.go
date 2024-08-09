package eks

import (
	"fmt"

	irsav1alpha1 "github.com/kkb0318/irsa-manager/api/v1alpha1"
)

func Validate(obj *irsav1alpha1.IRSASetup) error {
	if obj.Spec.IamOIDCProvider == "" {
		return fmt.Errorf("IamOIDCProvider parameter must be set when Mode is 'eks'")
	}
	return nil
}
