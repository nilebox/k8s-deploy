package release

import (
	"log"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api"
	"k8s.io/client-go/rest"

	deployv1 "github.com/nilebox/k8s-deploy/pkg/apis/v1"
	strategy "github.com/nilebox/k8s-deploy/pkg/release/strategy"
)

// ReleaseEventHandler can handle notifications for events to a Release resource
type ReleaseEventHandler struct {
	client *rest.RESTClient

	canary *strategy.Canary
}

func NewHandler(client *rest.RESTClient, clientset kubernetes.Interface) *ReleaseEventHandler {
	return &ReleaseEventHandler{
		client: client,
		canary: &strategy.Canary{
			Clientset: clientset,
		},
	}
}

func (h *ReleaseEventHandler) OnAdd(obj interface{}) {
	release := obj.(*deployv1.Release)
	log.Printf("[HANDLER] OnAdd %s", release.Metadata.SelfLink)

	if release.Metadata.Name == "" {
		log.Printf("ERROR Release name is empty!")
		return
	}
	h.handle(release)
}

func (h *ReleaseEventHandler) OnUpdate(oldObj, newObj interface{}) {
	oldRelease := oldObj.(*deployv1.Release)
	newRelease := newObj.(*deployv1.Release)
	log.Printf("[HANDLER] OnUpdate oldObj: %s", oldRelease.Metadata.SelfLink)
	log.Printf("[HANDLER] OnUpdate newObj: %s", newRelease.Metadata.SelfLink)
}

func (h *ReleaseEventHandler) OnDelete(obj interface{}) {
	release := obj.(*deployv1.Release)
	log.Printf("[HANDLER] OnDelete %s", release.Metadata.SelfLink)
}

func (h *ReleaseEventHandler) handle(release *deployv1.Release) {
	log.Printf("Processing new release %s", release.Metadata.Name)
	var err error
	switch release.Spec.Strategy.Type {
	case "Canary":
		log.Printf("Starting Canary deployment")
		err = h.canary.Run(release)
	case "BlueGreen":
		log.Printf("Starting BlueGreen deployment")
	case "":
		log.Printf("Starting default deployment (Canary)")
		err = h.canary.Run(release)
	default:
		log.Printf("Unknown deployment strategy: %s", release.Spec.Strategy.Type)
		return
	}

	if err != nil {
		log.Printf("[HANDLER] ERROR Failed to process Release object: ")
		return
	}

	// TODO: Use Patch instead of Put? At least create a deep copy before update

	// Update the status to signal that the Release has finished it's update and ready
	// releasePatch := deployv1.Release{
	// 	Status: deployv1.ReleaseStatus{
	// 		State: deployv1.ReleaseStateReady,
	// 	},
	// }
	release.Status.State = deployv1.ReleaseStateReady

	var result deployv1.Release
	// err = h.client.Patch(types.MergePatchType).
	// 	Resource(deployv1.ReleaseResourcePath).
	// 	Namespace(api.NamespaceDefault).
	// 	Name(release.Metadata.Name).
	// 	Body(release).
	// 	Do().Into(&result)
	err = h.client.Put().
		Resource(deployv1.ReleaseResourcePath).
		Namespace(api.NamespaceDefault).
		Name(release.Metadata.Name).
		Body(release).
		Do().Into(&result)

	if err != nil {
		panic(err)
	}
	log.Printf("PATCHED: %#v\n", result)

}
