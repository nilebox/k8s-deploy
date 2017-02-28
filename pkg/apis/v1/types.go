package v1

import (
	"encoding/json"

	"k8s.io/client-go/pkg/api"
	"k8s.io/client-go/pkg/api/meta"
	"k8s.io/client-go/pkg/api/unversioned"
	v1beta1 "k8s.io/client-go/pkg/apis/extensions/v1beta1"
	"k8s.io/client-go/pkg/util/intstr"
)

const (
	DeploymentDomain              = "k8s-deploy.atlassian.com"
	DeploymentResourceDescription = "Custom deployments support (Canary, Blue-green)"
	DeploymentResourceGroup       = DeploymentDomain

	DeploymentResourcePath         = "deployments"
	DeploymentResourceName         = "deployment." + DeploymentDomain
	DeploymentResourceVersion      = "v1"
	DeploymentResourceKind         = "Deployment"
	DeploymentResourceGroupVersion = DeploymentResourceGroup + "/" + DeploymentResourceVersion
)

// Deployment enables declarative updates for Pods and ReplicaSets.
type Deployment struct {
	unversioned.TypeMeta `json:",inline"`
	// Standard object metadata.
	// +optional
	api.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	// Specification of the desired behavior of the Deployment.
	// +optional
	Spec DeploymentSpec `json:"spec,omitempty" protobuf:"bytes,2,opt,name=spec"`

	// Most recently observed status of the Deployment.
	// +optional
	Status v1beta1.DeploymentStatus `json:"status,omitempty" protobuf:"bytes,3,opt,name=status"`
}

type DeploymentSpec struct {
	// Number of desired pods. This is a pointer to distinguish between explicit
	// zero and not specified. Defaults to 1.
	// +optional
	Replicas int32

	// Label selector for pods. Existing ReplicaSets whose pods are
	// selected by this will be the ones affected by this deployment.
	// +optional
	Selector *unversioned.LabelSelector

	// Template describes the pods that will be created.
	Template api.PodTemplateSpec

	// The deployment strategy to use to replace existing pods with new ones.
	// +optional
	Strategy DeploymentStrategy

	// Minimum number of seconds for which a newly created pod should be ready
	// without any of its container crashing, for it to be considered available.
	// Defaults to 0 (pod will be considered available as soon as it is ready)
	// +optional
	MinReadySeconds int32

	// The number of old ReplicaSets to retain to allow rollback.
	// This is a pointer to distinguish between explicit zero and not specified.
	// +optional
	RevisionHistoryLimit *int32

	// Indicates that the deployment is paused and will not be processed by the
	// deployment controller.
	// +optional
	Paused bool

	// The config this deployment is rolling back to. Will be cleared after rollback is done.
	// +optional
	RollbackTo *RollbackConfig

	// The maximum time in seconds for a deployment to make progress before it
	// is considered to be failed. The deployment controller will continue to
	// process failed deployments and a condition with a ProgressDeadlineExceeded
	// reason will be surfaced in the deployment status. Once autoRollback is
	// implemented, the deployment controller will automatically rollback failed
	// deployments. Note that progress will not be estimated during the time a
	// deployment is paused. This is not set by default.
	ProgressDeadlineSeconds *int32
}

type DeploymentStrategy struct {
	// Type of deployment. Can be "Recreate" or "RollingUpdate". Default is RollingUpdate.
	// +optional
	Type DeploymentStrategyType

	// Rolling update config params. Present only if DeploymentStrategyType =
	// RollingUpdate.
	//---
	// TODO: Update this to follow our convention for oneOf, whatever we decide it
	// to be.
	// +optional
	Canary *CanaryDeployment
}

type DeploymentStrategyType string

const (
	// Manage two native Deployment objects with Canary strategy.
	CanaryDeploymentStrategyType DeploymentStrategyType = "Canary"

	// Manage two native Deployment objects with Blue-green strategy.
	BlueGreenDeploymentStrategyType DeploymentStrategyType = "BlueGreen"
)

// Spec to control the desired behavior of rolling update.
type CanaryDeployment struct {
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

// DeploymentList is a list of Deployments.
type DeploymentList struct {
	unversioned.TypeMeta `json:",inline"`
	// Standard list metadata.
	// +optional
	unversioned.ListMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	// Items is the list of Deployments.
	Items []Deployment `json:"items" protobuf:"bytes,2,rep,name=items"`
}

// Required to satisfy Object interface
func (e *Deployment) GetObjectKind() unversioned.ObjectKind {
	return &e.TypeMeta
}

// Required to satisfy ObjectMetaAccessor interface
func (e *Deployment) GetObjectMeta() meta.Object {
	return &e.ObjectMeta
}

// Required to satisfy Object interface
func (el *DeploymentList) GetObjectKind() unversioned.ObjectKind {
	return &el.TypeMeta
}

// Required to satisfy ListMetaAccessor interface
func (el *DeploymentList) GetListMeta() unversioned.List {
	return &el.ListMeta
}

// The code below is used only to work around a known problem with third-party
// resources and ugorji. If/when these issues are resolved, the code below
// should no longer be required.

type DeploymentListCopy DeploymentList
type DeploymentCopy Deployment

func (e *Deployment) UnmarshalJSON(data []byte) error {
	tmp := DeploymentCopy{}
	err := json.Unmarshal(data, &tmp)
	if err != nil {
		return err
	}
	tmp2 := Deployment(tmp)
	*e = tmp2
	return nil
}

func (el *DeploymentList) UnmarshalJSON(data []byte) error {
	tmp := DeploymentListCopy{}
	err := json.Unmarshal(data, &tmp)
	if err != nil {
		return err
	}
	tmp2 := DeploymentList(tmp)
	*el = tmp2
	return nil
}
