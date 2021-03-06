package app

import (
	"context"
	"errors"
	"log"

	deployv1 "github.com/nilebox/k8s-deploy/pkg/apis/v1"
	"github.com/nilebox/k8s-deploy/pkg/client"
	"github.com/nilebox/k8s-deploy/pkg/release"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	fields "k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api"
	"k8s.io/client-go/pkg/apis/extensions/v1beta1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
)

type Server struct {
	RestConfig *rest.Config
}

func (s *Server) Run(ctx context.Context) error {
	log.Printf("Run()\n")

	log.Printf("Initializing REST client")
	clientset, err := kubernetes.NewForConfig(s.RestConfig)
	if err != nil {
		return err
	}

	log.Printf("Initializing Release client")
	releaseClient, releaseScheme, err := client.NewClient(s.RestConfig)
	if err != nil {
		return err
	}

	log.Printf("Ensure ThirdPartyResource Release exists")
	// Ensure ThirdPartyResource Release exists
	err = ensureReleaseResourceExists(clientset)
	if err != nil {
		// TODO retry
		log.Printf("Failed to create resource %s: %v", deployv1.ReleaseResourceName, err)
		return err
	}

	log.Printf("Watch Release objects")
	// Watch Release objects
	handler := release.NewHandler(releaseClient, clientset)
	releaseInformer, err := watchReleases(ctx, releaseClient, releaseScheme, handler)
	if err != nil {
		log.Printf("Failed to register watch for Release resource: %v", err)
		return err
	}

	log.Printf("WaitForCacheSync")
	// We must wait for releaseInformer to populate its cache to avoid reading from an empty cache
	// in case of resource-generated evxents.
	if !cache.WaitForCacheSync(ctx.Done(), releaseInformer.HasSynced) {
		return errors.New("wait for Release Informer was cancelled")
	}

	<-ctx.Done()
	return ctx.Err()
}

func ensureReleaseResourceExists(clientset kubernetes.Interface) error {
	// initialize third party resource if it does not exist
	tpr, err := clientset.ExtensionsV1beta1().ThirdPartyResources().Get(deployv1.ReleaseResourceName, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			log.Printf("NOT FOUND: releases TPR\n")

			tpr := &v1beta1.ThirdPartyResource{
				ObjectMeta: metav1.ObjectMeta{
					Name: deployv1.ReleaseResourceName,
					// This is for Smith support https://github.com/atlassian/smith/blob/master/docs/design/managing-resources.md
					Annotations: map[string]string{
						"smith.atlassian.com/TprReadyWhenFieldPath":  "status.state",
						"smith.atlassian.com/TprReadyWhenFieldValue": "Ready",
					},
				},
				Versions: []v1beta1.APIVersion{
					{Name: deployv1.ReleaseResourceVersion},
				},
				Description: deployv1.ReleaseResourceDescription,
			}

			result, err := clientset.ExtensionsV1beta1().ThirdPartyResources().Create(tpr)
			if err != nil {
				return err
			}
			log.Printf("CREATED: %#v\nFROM: %#v\n", result, tpr)
		} else {
			return err
		}
	} else {
		log.Printf("SKIPPING: already exists %s", tpr.ObjectMeta.SelfLink)
	}
	return nil
}

func watchReleases(ctx context.Context, releaseClient cache.Getter, releaseScheme *runtime.Scheme, handler *release.ReleaseEventHandler) (cache.Controller, error) {
	parameterCodec := runtime.NewParameterCodec(releaseScheme)

	source := newListWatchFromClient(
		releaseClient,
		deployv1.ReleaseResourcePath,
		api.NamespaceAll,
		fields.Everything(),
		parameterCodec)

	store, controller := cache.NewInformer(
		source,

		// The object type.
		&deployv1.Release{},

		// resyncPeriod
		// Every resyncPeriod, all resources in the cache will retrigger events.
		// Set to 0 to disable the resync.
		//time.Second*10,
		0,

		// Your custom resource event handlers.
		handler)

	// store can be used to List and Get
	// NEVER modify objects from the store. It's a read-only, local cache.
	log.Println("listing releases from store:")
	for _, obj := range store.List() {
		release := obj.(*deployv1.Release)

		// This will likely be empty the first run, but may not
		log.Printf("Existing release: %#v\n", release)
	}

	go controller.Run(ctx.Done())

	return controller, nil
}

// newListWatchFromClient is a copy of cache.NewListWatchFromClient() method with custom codec
// Cannot use cache.NewListWatchFromClient() because it uses global api.ParameterCodec which uses global
// api.Scheme which does not know about custom types (Release in our case) group/version.
// cache.NewListWatchFromClient(releaseClient, deployv1.ReleaseResourcePath, apiv1.NamespaceAll, fields.Everything())
func newListWatchFromClient(c cache.Getter, resource string, namespace string, fieldSelector fields.Selector, paramCodec runtime.ParameterCodec) *cache.ListWatch {
	listFunc := func(options metav1.ListOptions) (runtime.Object, error) {
		return c.Get().
			Namespace(namespace).
			Resource(resource).
			VersionedParams(&options, paramCodec).
			FieldsSelectorParam(fieldSelector).
			Do().
			Get()
	}
	watchFunc := func(options metav1.ListOptions) (watch.Interface, error) {
		return c.Get().
			Prefix("watch").
			Namespace(namespace).
			Resource(resource).
			VersionedParams(&options, paramCodec).
			FieldsSelectorParam(fieldSelector).
			Watch()
	}
	return &cache.ListWatch{ListFunc: listFunc, WatchFunc: watchFunc}
}
