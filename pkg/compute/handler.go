package compute

import (
	"log"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api"
	"k8s.io/client-go/rest"

	deployv1 "github.com/nilebox/k8s-deploy/pkg/apis/v1"
	strategy "github.com/nilebox/k8s-deploy/pkg/compute/strategy"
)

// ComputeEventHandler can handle notifications for events to a Compute resource
type ComputeEventHandler struct {
	client *rest.RESTClient

	canary *strategy.Canary
}

func NewHandler(client *rest.RESTClient, clientset kubernetes.Interface) *ComputeEventHandler {
	return &ComputeEventHandler{
		client: client,
		canary: &strategy.Canary{
			Clientset: clientset,
		},
	}
}

func (h *ComputeEventHandler) OnAdd(obj interface{}) {
	compute := obj.(*deployv1.Compute)
	log.Printf("[HANDLER] OnAdd %s", compute.Metadata.SelfLink)

	if compute.Metadata.Name == "" {
		log.Printf("ERROR Compute name is empty!")
		return
	}
	h.handle(compute)
}

func (h *ComputeEventHandler) OnUpdate(oldObj, newObj interface{}) {
	oldCompute := oldObj.(*deployv1.Compute)
	newCompute := newObj.(*deployv1.Compute)
	log.Printf("[HANDLER] OnUpdate oldObj: %s", oldCompute.Metadata.SelfLink)
	log.Printf("[HANDLER] OnUpdate newObj: %s", newCompute.Metadata.SelfLink)
}

func (h *ComputeEventHandler) OnDelete(obj interface{}) {
	compute := obj.(*deployv1.Compute)
	log.Printf("[HANDLER] OnDelete %s", compute.Metadata.SelfLink)
}

func (h *ComputeEventHandler) handle(compute *deployv1.Compute) {
	log.Printf("Processing new compute %s", compute.Metadata.Name)
	var err error
	switch compute.Spec.Strategy.Type {
	case "Canary":
		log.Printf("Starting Canary deployment")
		err = h.canary.Run(compute)
	case "BlueGreen":
		log.Printf("Starting BlueGreen deployment")
	case "":
		log.Printf("Starting default deployment (Canary)")
		err = h.canary.Run(compute)
	default:
		log.Printf("Unknown deployment strategy: %s", compute.Spec.Strategy.Type)
		return
	}

	if err != nil {
		log.Printf("[HANDLER] ERROR Failed to process Compute object: ")
		return
	}

	// TODO: Use Patch instead of Put? At least create a deep copy before update

	// Update the status to signal that the Compute has finished it's update and ready
	// computePatch := deployv1.Compute{
	// 	Status: deployv1.ComputeStatus{
	// 		State: deployv1.ComputeStateReady,
	// 	},
	// }
	compute.Status.State = deployv1.ComputeStateReady

	var result deployv1.Compute
	// err = h.client.Patch(types.MergePatchType).
	// 	Resource(deployv1.ComputeResourcePath).
	// 	Namespace(api.NamespaceDefault).
	// 	Name(compute.Metadata.Name).
	// 	Body(compute).
	// 	Do().Into(&result)
	err = h.client.Put().
		Resource(deployv1.ComputeResourcePath).
		Namespace(api.NamespaceDefault).
		Name(compute.Metadata.Name).
		Body(compute).
		Do().Into(&result)

	if err != nil {
		panic(err)
	}
	log.Printf("PATCHED: %#v\n", result)

}
