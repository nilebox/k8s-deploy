package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

var _ runtime.Object = &ComputeList{}
var _ metav1.ListMetaAccessor = &ComputeList{}

var _ runtime.Object = &Compute{}
var _ metav1.ObjectMetaAccessor = &Compute{}
