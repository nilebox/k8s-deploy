package app

import (
	"context"
	"errors"
	"log"
	"time"

	deployv1 "github.com/nilebox/k8s-deploy/pkg/apis/v1"
	"github.com/nilebox/k8s-deploy/pkg/client"
	"github.com/nilebox/k8s-deploy/pkg/release"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api"
	apierrors "k8s.io/client-go/pkg/api/errors"
	apiv1 "k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/pkg/apis/extensions/v1beta1"
	fields "k8s.io/client-go/pkg/fields"
	"k8s.io/client-go/pkg/runtime"
	"k8s.io/client-go/pkg/watch"
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
	handler := release.NewHandler(clientset)
	releaseInformer, err := watchReleases(ctx, releaseClient, releaseScheme, handler)
	if err != nil {
		log.Printf("Failed to register watch for Release resource: %v", err)
		return err
	}

	log.Printf("WaitForCacheSync")
	// We must wait for tmplInf to populate its cache to avoid reading from an empty cache
	// in case of resource-generated evxents.
	if !cache.WaitForCacheSync(ctx.Done(), releaseInformer.HasSynced) {
		return errors.New("wait for Release Informer was cancelled")
	}

	<-ctx.Done()
	return ctx.Err()
}

func ensureReleaseResourceExists(clientset kubernetes.Interface) error {
	// initialize third party resource if it does not exist
	tpr, err := clientset.ExtensionsV1beta1().ThirdPartyResources().Get(deployv1.ReleaseResourceName)
	if err != nil {
		if apierrors.IsNotFound(err) {
			log.Printf("NOT FOUND: releases TPR\n")

			tpr := &v1beta1.ThirdPartyResource{
				ObjectMeta: apiv1.ObjectMeta{
					Name: deployv1.ReleaseResourceName,
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

func watchReleases(ctx context.Context, releaseClient cache.Getter, releaseScheme *runtime.Scheme, handler *release.ReleaseEventHandler) (*cache.Controller, error) {
	parameterCodec := runtime.NewParameterCodec(releaseScheme)

	// // Cannot use cache.NewListWatchFromClient() because it uses global api.ParameterCodec which uses global
	// // api.Scheme which does not know about Release group/version.
	// // cache.NewListWatchFromClient(releaseClient, deployv1.ReleaseResourcePath, apiv1.NamespaceAll, fields.Everything())
	// releaseInformer := cache.NewSharedInformer(&cache.ListWatch{
	// 	ListFunc: func(options api.ListOptions) (runtime.Object, error) {
	// 		return releaseClient.Get().
	// 			Resource(deployv1.ReleaseResourcePath).
	// 			VersionedParams(&options, parameterCodec).
	// 			Do().
	// 			Get()
	// 	},
	// 	WatchFunc: func(options api.ListOptions) (watch.Interface, error) {
	// 		return releaseClient.Get().
	// 			Prefix("watch").
	// 			Resource(deployv1.ReleaseResourcePath).
	// 			VersionedParams(&options, parameterCodec).
	// 			Watch()
	// 	},
	// }, &deployv1.Release{}, 0)

	// if err := releaseInformer.AddEventHandler(handler); err != nil {
	// 	return nil, err
	// }

	// go releaseInformer.Run(ctx.Done())

	// return releaseInformer, nil

	var s fields.Selector
	s = fields.Everything()
	source := newListWatchFromClient(
		releaseClient,
		deployv1.ReleaseResourcePath,
		api.NamespaceAll,
		s,
		parameterCodec)

	store, controller := cache.NewInformer(
		source,

		// The object type.
		&deployv1.Release{},

		// resyncPeriod
		// Every resyncPeriod, all resources in the cache will retrigger events.
		// Set to 0 to disable the resync.
		time.Second*10,

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
func newListWatchFromClient(c cache.Getter, resource string, namespace string, fieldSelector fields.Selector, paramCodec runtime.ParameterCodec) *cache.ListWatch {
	listFunc := func(options api.ListOptions) (runtime.Object, error) {
		return c.Get().
			Namespace(namespace).
			Resource(resource).
			VersionedParams(&options, paramCodec).
			FieldsSelectorParam(fieldSelector).
			Do().
			Get()
	}
	watchFunc := func(options api.ListOptions) (watch.Interface, error) {
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
