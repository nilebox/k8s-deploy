// +build integration

package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"testing"

	"k8s.io/client-go/rest"
	// deployv1 "github.com/nilebox/k8s-deploy/pkg/apis/v1"
	// "github.com/nilebox/k8s-deploy/pkg/client"
	// "k8s.io/client-go/kubernetes"
	// "k8s.io/client-go/pkg/api"
	// "k8s.io/client-go/pkg/api/errors"
	// "k8s.io/client-go/pkg/api/v1"
	// "k8s.io/client-go/pkg/apis/extensions/v1beta1"
	// "k8s.io/client-go/rest"
)

func TestCanaryRelease(t *testing.T) {
	fmt.Printf("Start\n")
	config := configFromEnv(t)

	fmt.Printf("Setup context\n")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	fmt.Printf("Run\n")
	runWithConfig(ctx, config)
	fmt.Printf("Finish\n")

	// clientset, err := kubernetes.NewForConfig(config)

	// // initialize third party resource if it does not exist
	// tpr, err := clientset.ExtensionsV1beta1().ThirdPartyResources().Get(deployv1.ReleaseResourceName)
	// if err != nil {
	// 	if errors.IsNotFound(err) {
	// 		fmt.Printf("NOT FOUND: releases TPR\n")

	// 		tpr := &v1beta1.ThirdPartyResource{
	// 			ObjectMeta: v1.ObjectMeta{
	// 				Name: deployv1.ReleaseResourceName,
	// 			},
	// 			Versions: []v1beta1.APIVersion{
	// 				{Name: deployv1.ReleaseResourceVersion},
	// 			},
	// 			Description: deployv1.ReleaseResourceDescription,
	// 		}

	// 		result, err := clientset.ExtensionsV1beta1().ThirdPartyResources().Create(tpr)
	// 		if err != nil {
	// 			panic(err)
	// 		}
	// 		fmt.Printf("CREATED: %#v\nFROM: %#v\n", result, tpr)

	// 		time.Sleep(10 * time.Second) // Wait until the TPR initialization is finished
	// 	} else {
	// 		panic(err)
	// 	}
	// } else {
	// 	fmt.Printf("SKIPPING: already exists %#v\n", tpr)
	// }

	// tprclient, _, err := client.NewClient(config)

	// if err != nil {
	// 	panic(err)
	// }

	// var release deployv1.Release

	// err = tprclient.Get().
	// 	Resource(deployv1.ReleaseResourcePath).
	// 	Namespace(api.NamespaceDefault).
	// 	Name("release1").
	// 	Do().Into(&release)

	// if err != nil {
	// 	if errors.IsNotFound(err) {
	// 		fmt.Printf("NOT FOUND: releases TPR instance\n")

	// 		// Create an instance of our TPR
	// 		release := &deployv1.Release{
	// 			ObjectMeta: api.ObjectMeta{
	// 				Name: "release1",
	// 			},
	// 			Spec: deployv1.ReleaseSpec{
	// 				Replicas: 3,
	// 			},
	// 		}

	// 		var result deployv1.Release
	// 		err = tprclient.Post().
	// 			Resource(deployv1.ReleaseResourcePath).
	// 			Namespace(api.NamespaceDefault).
	// 			Body(release).
	// 			Do().Into(&result)

	// 		if err != nil {
	// 			panic(err)
	// 		}
	// 		fmt.Printf("CREATED: %#v\n", result)
	// 	} else {
	// 		panic(err)
	// 	}
	// } else {
	// 	fmt.Printf("GET: %#v\n", release)
	// }

	// // Fetch a list of our TPRs
	// releaseList := deployv1.ReleaseList{}
	// err = tprclient.Get().Resource(deployv1.ReleaseResourcePath).Do().Into(&releaseList)
	// if err != nil {
	// 	panic(err)
	// }
	// fmt.Printf("LIST: %#v\n", releaseList)
}

func configFromEnv(t *testing.T) *rest.Config {
	host, port := os.Getenv("KUBERNETES_SERVICE_HOST"), os.Getenv("KUBERNETES_SERVICE_PORT")
	if len(host) == 0 || len(port) == 0 {
		t.Fatal("Unable to load cluster configuration, KUBERNETES_SERVICE_HOST and KUBERNETES_SERVICE_PORT must be defined")
	}
	return &rest.Config{
		Host: "https://" + net.JoinHostPort(host, port),
		TLSClientConfig: rest.TLSClientConfig{
			CAFile:   os.Getenv("KUBERNETES_CA_PATH"),
			CertFile: os.Getenv("KUBERNETES_CLIENT_CERT"),
			KeyFile:  os.Getenv("KUBERNETES_CLIENT_KEY"),
		},
	}
}
