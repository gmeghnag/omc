package v2

import (
	"sigs.k8s.io/controller-runtime/pkg/conversion"
)

var _ conversion.Hub = (*ServiceMeshControlPlane)(nil)

// Hub marks v2 SMCP resource as the storage version
func (smcp *ServiceMeshControlPlane) Hub() {}
