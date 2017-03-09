package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

var _ runtime.Object = &ReleaseList{}
var _ metav1.ListMetaAccessor = &ReleaseList{}

var _ runtime.Object = &Release{}
var _ metav1.ObjectMetaAccessor = &Release{}
