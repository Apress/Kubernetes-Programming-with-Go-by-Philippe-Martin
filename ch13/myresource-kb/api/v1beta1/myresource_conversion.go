package v1beta1

import (
	"github.com/myid/myresource/api/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/conversion"
)

func (src *MyResource) ConvertTo(
	dstRaw conversion.Hub,
) error {
	dst := dstRaw.(*v1alpha1.MyResource)
	dst.Spec.Memory = src.Spec.MemoryRequest
	// Copy other fields
	dst.ObjectMeta = src.ObjectMeta
	dst.Spec.Image = src.Spec.Image
	dst.Status.State = src.Status.State
	return nil
}

func (dst *MyResource) ConvertFrom(
	srcRaw conversion.Hub,
) error {
	src := srcRaw.(*v1alpha1.MyResource)
	dst.Spec.MemoryRequest = src.Spec.Memory
	// Copy other fields
	dst.ObjectMeta = src.ObjectMeta
	dst.Spec.Image = src.Spec.Image
	dst.Status.State = src.Status.State
	return nil
}
