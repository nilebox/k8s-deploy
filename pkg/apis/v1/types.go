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
	ReleaseDomain              = "deploy.k8s"
	ReleaseResourceDescription = "Custom releases support (Canary, Blue-green)"
	ReleaseResourceGroup       = ReleaseDomain

	ReleaseResourcePath         = "releases"
	ReleaseResourceName         = "release." + ReleaseDomain
	ReleaseResourceVersion      = "v1"
	ReleaseResourceKind         = "Release"
	ReleaseResourceGroupVersion = ReleaseResourceGroup + "/" + ReleaseResourceVersion
)

// Release enables declarative updates for Pods and ReplicaSets.
type Release struct {
	metav1.TypeMeta `json:",inline"`

	// *** SECRET KNOWLEDGE ***: Don't call the field below ObjectMeta, it will blow up the JSON deserialization
	// Issue: https://github.com/kubernetes/client-go/issues/8

	// Standard object metadata.
	// +optional
	Metadata apiv1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	// Specification of the desired behavior of the Release.
	// +optional
	Spec ReleaseSpec `json:"spec,omitempty" protobuf:"bytes,2,opt,name=spec"`

	// Most recently observed status of the Release.
	// +optional
	Status ReleaseStatus `json:"status,omitempty" protobuf:"bytes,3,opt,name=status"`
}

type ReleaseStatus struct {
	// State is the current state of the Release
	State ReleaseState
}

type ReleaseState string

// These are valid release state
const (
	ReleaseStateReady   ReleaseState = "Ready"
	ReleaseStateFailure ReleaseState = "Failure"
)

type ReleaseSpec struct {
	// Number of desired pods. This is a pointer to distinguish between explicit
	// zero and not specified. Defaults to 1.
	// +optional
	Replicas *int32

	// Label selector for pods. Existing ReplicaSets whose pods are
	// selected by this will be the ones affected by this release.
	// +optional
	Selector *metav1.LabelSelector

	// Template describes the pods that will be created.
	Template apiv1.PodTemplateSpec

	// The release strategy to use to replace existing pods with new ones.
	// +optional
	Strategy ReleaseStrategy

	// Minimum number of seconds for which a newly created pod should be ready
	// without any of its container crashing, for it to be considered available.
	// Defaults to 0 (pod will be considered available as soon as it is ready)
	// +optional
	MinReadySeconds int32

	// The number of old ReplicaSets to retain to allow rollback.
	// This is a pointer to distinguish between explicit zero and not specified.
	// +optional
	RevisionHistoryLimit *int32

	// Indicates that the release is paused and will not be processed by the
	// release controller.
	// +optional
	Paused bool

	// The config this release is rolling back to. Will be cleared after rollback is done.
	// +optional
	RollbackTo *RollbackConfig

	// The maximum time in seconds for a release to make progress before it
	// is considered to be failed. The release controller will continue to
	// process failed releases and a condition with a ProgressDeadlineExceeded
	// reason will be surfaced in the release status. Once autoRollback is
	// implemented, the release controller will automatically rollback failed
	// releases. Note that progress will not be estimated during the time a
	// release is paused. This is not set by default.
	ProgressDeadlineSeconds *int32
}

type ReleaseStrategy struct {
	// Type of release. Can be "Recreate" or "RollingUpdate". Default is RollingUpdate.
	// +optional
	Type ReleaseStrategyType

	// Rolling update config params. Present only if ReleaseStrategyType =
	// RollingUpdate.
	//---
	// TODO: Update this to follow our convention for oneOf, whatever we decide it
	// to be.
	// +optional
	Canary *CanaryRelease
}

type ReleaseStrategyType string

const (
	// Manage two native Release objects with Canary strategy.
	CanaryReleaseStrategyType ReleaseStrategyType = "Canary"

	// Manage two native Release objects with Blue-green strategy.
	BlueGreenReleaseStrategyType ReleaseStrategyType = "BlueGreen"
)

// Spec to control the desired behavior of rolling update.
type CanaryRelease struct {
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

// ReleaseList is a list of Releases.
type ReleaseList struct {
	metav1.TypeMeta `json:",inline"`

	// *** SECRET KNOWLEDGE ***: Don't call the field below ListMeta, it will blow up the JSON deserialization
	// Issue: https://github.com/kubernetes/client-go/issues/8

	// Standard list metadata.
	// +optional
	Metadata metav1.ListMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	// Items is the list of Releases.
	Items []Release `json:"items" protobuf:"bytes,2,rep,name=items"`
}

// Required to satisfy Object interface
func (e *Release) GetObjectKind() schema.ObjectKind {
	return &e.TypeMeta
}

// Required to satisfy ObjectMetaAccessor interface
func (e *Release) GetObjectMeta() metav1.Object {
	return &e.Metadata
}

// Required to satisfy Object interface
func (el *ReleaseList) GetObjectKind() schema.ObjectKind {
	return &el.TypeMeta
}

// Required to satisfy ListMetaAccessor interface
func (el *ReleaseList) GetListMeta() metav1.List {
	return &el.Metadata
}

// The code below is used only to work around a known problem with third-party
// resources and ugorji. If/when these issues are resolved, the code below
// should no longer be required.

type ReleaseListCopy ReleaseList
type ReleaseCopy Release

func (e *Release) UnmarshalJSON(data []byte) error {
	tmp := ReleaseCopy{}
	err := json.Unmarshal(data, &tmp)
	if err != nil {
		log.Printf("UnmarshalJSON Error")
		return err
	}
	log.Printf("UnmarshalJSON: %s", tmp.Metadata.SelfLink)
	tmp2 := Release(tmp)
	*e = tmp2
	return nil
}

func (el *ReleaseList) UnmarshalJSON(data []byte) error {
	tmp := ReleaseListCopy{}
	err := json.Unmarshal(data, &tmp)
	if err != nil {
		log.Printf("UnmarshalJSON Error")
		return err
	}
	log.Printf("UnmarshalJSON: %s", tmp.Metadata.SelfLink)
	tmp2 := ReleaseList(tmp)
	*el = tmp2
	return nil
}

func (e *Release) UnmarshalText(data []byte) error {
	log.Printf("UnmarshalText")
	return json.Unmarshal(data, e)
}

func (el *ReleaseList) UnmarshalText(data []byte) error {
	log.Printf("UnmarshalText")
	return json.Unmarshal(data, el)
}
