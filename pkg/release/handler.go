package release

import (
	"log"

	"k8s.io/client-go/kubernetes"

	deployv1 "github.com/nilebox/k8s-deploy/pkg/apis/v1"
	strategy "github.com/nilebox/k8s-deploy/pkg/release/strategy"
)

// ReleaseEventHandler can handle notifications for events to a Release resource
type ReleaseEventHandler struct {
	canary *strategy.Canary
}

func NewHandler(clientset kubernetes.Interface) *ReleaseEventHandler {
	return &ReleaseEventHandler{
		canary: &strategy.Canary{
			Clientset: clientset,
		},
	}
}

func (h *ReleaseEventHandler) OnAdd(obj interface{}) {
	release := obj.(*deployv1.Release)
	log.Printf("[REH] OnAdd %s", release.Metadata.SelfLink)

	if release.TypeMeta.Kind == "" {
		log.Printf("ERROR Unknown release, skipping")
		return
	}
	h.handle(release)
}

func (h *ReleaseEventHandler) OnUpdate(oldObj, newObj interface{}) {
	oldRelease := oldObj.(*deployv1.Release)
	newRelease := newObj.(*deployv1.Release)
	log.Printf("[REH] OnUpdate oldObj: %s", oldRelease.Metadata.SelfLink)
	log.Printf("[REH] OnUpdate newObj: %s", newRelease.Metadata.SelfLink)
}

func (h *ReleaseEventHandler) OnDelete(obj interface{}) {
	release := obj.(*deployv1.Release)
	log.Printf("[REH] OnDelete %s", release.Metadata.SelfLink)
}

func (h *ReleaseEventHandler) handle(release *deployv1.Release) {
	log.Printf("Processing new release %s", release.Metadata.Name)
	switch release.Spec.Strategy.Type {
	case "Canary":
		log.Printf("Starting Canary deployment")
		h.canary.Run(release)
	case "BlueGreen":
		log.Printf("Starting BlueGreen deployment")
	case "":
		log.Printf("Starting default deployment (Canary)")
		h.canary.Run(release)
	default:
		log.Printf("Unknown deployment strategy: %s", release.Spec.Strategy.Type)
	}
}
