package v1

import (
	"encoding/json"
	"log"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/intstr"
	apiv1 "k8s.io/client-go/pkg/api/v1"
)

const (
	ComputeDomain              = "resource.atlassian.com"
	ComputeResourceDescription = "Custom computes support (Canary, Blue-green)"
	ComputeResourceGroup       = ComputeDomain

	ComputeResourcePath         = "computes"
	ComputeResourceName         = "compute." + ComputeDomain
	ComputeResourceVersion      = "v1"
	ComputeResourceKind         = "Compute"
	ComputeResourceGroupVersion = ComputeResourceGroup + "/" + ComputeResourceVersion
)

// Compute enables declarative updates for Pods and ReplicaSets.
type Compute struct {
	metav1.TypeMeta `json:",inline"`

	// *** SECRET KNOWLEDGE ***: Don't call the field below ObjectMeta, it will blow up the JSON deserialization
	// Issue: https://github.com/kubernetes/client-go/issues/8

	// Standard object metadata.
	// +optional
	Metadata apiv1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	// Specification of the desired behavior of the Compute.
	// +optional
	Spec ComputeSpec `json:"spec,omitempty" protobuf:"bytes,2,opt,name=spec"`

	// Most recently observed status of the Compute.
	// +optional
	Status ComputeStatus `json:"status,omitempty"`
}

type ComputeStatus struct {
	// State is the current state of the Compute
	State ComputeState `json:"state,omitempty"`
}

type ComputeState string

// These are valid compute state
const (
	ComputeStateReady   ComputeState = "Ready"
	ComputeStateFailure ComputeState = "Failure"
)

type ComputeSpec struct {
	// Number of desired pods. This is a pointer to distinguish between explicit
	// zero and not specified. Defaults to 1.
	// +optional
	Replicas *int32

	// Label selector for pods. Existing ReplicaSets whose pods are
	// selected by this will be the ones affected by this compute.
	// +optional
	Selector *metav1.LabelSelector

	// Template describes the pods that will be created.
	Template apiv1.PodTemplateSpec

	// The compute strategy to use to replace existing pods with new ones.
	// +optional
	Strategy ComputeStrategy

	// Minimum number of seconds for which a newly created pod should be ready
	// without any of its container crashing, for it to be considered available.
	// Defaults to 0 (pod will be considered available as soon as it is ready)
	// +optional
	MinReadySeconds int32

	// The number of old ReplicaSets to retain to allow rollback.
	// This is a pointer to distinguish between explicit zero and not specified.
	// +optional
	RevisionHistoryLimit *int32

	// Indicates that the compute is paused and will not be processed by the
	// compute controller.
	// +optional
	Paused bool

	// The config this compute is rolling back to. Will be cleared after rollback is done.
	// +optional
	RollbackTo *RollbackConfig

	// The maximum time in seconds for a compute to make progress before it
	// is considered to be failed. The compute controller will continue to
	// process failed computes and a condition with a ProgressDeadlineExceeded
	// reason will be surfaced in the compute status. Once autoRollback is
	// implemented, the compute controller will automatically rollback failed
	// computes. Note that progress will not be estimated during the time a
	// compute is paused. This is not set by default.
	ProgressDeadlineSeconds *int32
}

type ComputeStrategy struct {
	// Type of compute. Can be "Recreate" or "RollingUpdate". Default is RollingUpdate.
	// +optional
	Type ComputeStrategyType

	// Rolling update config params. Present only if ComputeStrategyType =
	// RollingUpdate.
	//---
	// TODO: Update this to follow our convention for oneOf, whatever we decide it
	// to be.
	// +optional
	Canary *CanaryCompute
}

type ComputeStrategyType string

const (
	// Manage two native Compute objects with Canary strategy.
	CanaryComputeStrategyType ComputeStrategyType = "Canary"

	// Manage two native Compute objects with Blue-green strategy.
	BlueGreenComputeStrategyType ComputeStrategyType = "BlueGreen"
)

// Spec to control the desired behavior of rolling update.
type CanaryCompute struct {
	// The maximum number of pods that can be unavailable during the update.
	// Value can be an absolute number (ex: 5) or a percentage of total pods at the start of update (ex: 10%).
	// Absolute number is calculated from percentage by rounding down.
	// This can not be 0 if MaxSurge is 0.
	// By default, a fixed value of 1 is used.
	// Example: when this is set to 30%, the old RC can be scaled down by 30%
	// immediately when the rolling update starts. Once new pods are ready, old RC
	// can be scaled down further, followed by scaling up the new RC, ensuring
	// that at least 70% of original number of pods are available at all times
	// during the update.
	// +optional
	MaxUnavailable intstr.IntOrString

	// The maximum number of pods that can be scheduled above the original number of
	// pods.
	// Value can be an absolute number (ex: 5) or a percentage of total pods at
	// the start of the update (ex: 10%). This can not be 0 if MaxUnavailable is 0.
	// Absolute number is calculated from percentage by rounding up.
	// By default, a value of 1 is used.
	// Example: when this is set to 30%, the new RC can be scaled up by 30%
	// immediately when the rolling update starts. Once old pods have been killed,
	// new RC can be scaled up further, ensuring that total number of pods running
	// at any time during the update is atmost 130% of original pods.
	// +optional
	MaxSurge intstr.IntOrString
}

type RollbackConfig struct {
	// The revision to rollback to. If set to 0, rollbck to the last revision.
	// +optional
	Revision int64
}

// ComputeList is a list of Computes.
type ComputeList struct {
	metav1.TypeMeta `json:",inline"`

	// *** SECRET KNOWLEDGE ***: Don't call the field below ListMeta, it will blow up the JSON deserialization
	// Issue: https://github.com/kubernetes/client-go/issues/8

	// Standard list metadata.
	// +optional
	Metadata metav1.ListMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	// Items is the list of Computes.
	Items []Compute `json:"items" protobuf:"bytes,2,rep,name=items"`
}

// Required to satisfy Object interface
func (e *Compute) GetObjectKind() schema.ObjectKind {
	return &e.TypeMeta
}

// Required to satisfy ObjectMetaAccessor interface
func (e *Compute) GetObjectMeta() metav1.Object {
	return &e.Metadata
}

// Required to satisfy Object interface
func (el *ComputeList) GetObjectKind() schema.ObjectKind {
	return &el.TypeMeta
}

// Required to satisfy ListMetaAccessor interface
func (el *ComputeList) GetListMeta() metav1.List {
	return &el.Metadata
}

// The code below is used only to work around a known problem with third-party
// resources and ugorji. If/when these issues are resolved, the code below
// should no longer be required.

type ComputeListCopy ComputeList
type ComputeCopy Compute

func (e *Compute) UnmarshalJSON(data []byte) error {
	tmp := ComputeCopy{}
	err := json.Unmarshal(data, &tmp)
	if err != nil {
		log.Printf("UnmarshalJSON Error")
		return err
	}
	log.Printf("UnmarshalJSON: %s", tmp.Metadata.SelfLink)
	tmp2 := Compute(tmp)
	*e = tmp2
	return nil
}

func (el *ComputeList) UnmarshalJSON(data []byte) error {
	tmp := ComputeListCopy{}
	err := json.Unmarshal(data, &tmp)
	if err != nil {
		log.Printf("UnmarshalJSON Error")
		return err
	}
	log.Printf("UnmarshalJSON: %s", tmp.Metadata.SelfLink)
	tmp2 := ComputeList(tmp)
	*el = tmp2
	return nil
}

func (e *Compute) UnmarshalText(data []byte) error {
	log.Printf("UnmarshalText")
	return json.Unmarshal(data, e)
}

func (el *ComputeList) UnmarshalText(data []byte) error {
	log.Printf("UnmarshalText")
	return json.Unmarshal(data, el)
}
