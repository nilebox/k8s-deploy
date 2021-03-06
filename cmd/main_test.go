// +build integration

package main

import (
	"context"
	"log"
	"testing"

	"github.com/nilebox/k8s-deploy/pkg/client"

	"time"

	deployv1 "github.com/nilebox/k8s-deploy/pkg/apis/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/pkg/api"
	apiv1 "k8s.io/client-go/pkg/api/v1"
)

func TestCanaryRelease(t *testing.T) {
	log.Printf("Start\n")
	config := configFromEnv()

	log.Printf("Setup context\n")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	log.Printf("Run\n")
	go func() {
		if err := runWithConfig(ctx, config); err != context.Canceled && err != context.DeadlineExceeded {
			panic(err)
		}
	}()
	log.Printf("Finish\n")

	time.Sleep(5 * time.Second) // Wait until the app starts and creates the Release TPR

	tprclient, _, err := client.NewClient(config)

	if err != nil {
		panic(err)
	}

	var release deployv1.Release

	err = tprclient.Get().
		Resource(deployv1.ReleaseResourcePath).
		Namespace(api.NamespaceDefault).
		Name("release2").
		Do().Into(&release)

	if err != nil {
		if errors.IsNotFound(err) {
			log.Printf("NOT FOUND: release2 instance\n")

			replicas := int32(3)
			// Create an instance of our TPR
			release := &deployv1.Release{
				TypeMeta: metav1.TypeMeta{
					APIVersion: deployv1.ReleaseResourceGroupVersion,
					Kind:       deployv1.ReleaseResourceKind,
				},
				Metadata: apiv1.ObjectMeta{
					Name: "release2",
				},
				Spec: deployv1.ReleaseSpec{
					Replicas: &replicas,
					Selector: &metav1.LabelSelector{
						MatchLabels: map[string]string{
							"app":     "k8s-deploy-test",
							"version": "1.0",
						},
					},
					Template: apiv1.PodTemplateSpec{
						ObjectMeta: apiv1.ObjectMeta{
							Labels: map[string]string{
								"app":     "k8s-deploy-test",
								"version": "1.0",
							},
						},
						Spec: apiv1.PodSpec{
							Containers: []apiv1.Container{
								apiv1.Container{
									Name:  "tea",
									Image: "nginxdemos/hello",
									Ports: []apiv1.ContainerPort{
										apiv1.ContainerPort{
											ContainerPort: 80,
										},
									},
								},
							},
						},
					},
				},
			}

			var result deployv1.Release
			err = tprclient.Post().
				Resource(deployv1.ReleaseResourcePath).
				Namespace(api.NamespaceDefault).
				Body(release).
				Do().Into(&result)

			if err != nil {
				panic(err)
			}
			log.Printf("CREATED: %#v\n", result)
		} else {
			panic(err)
		}
	} else {
		log.Printf("GET: %#v\n", release)
	}

	// // Fetch a list of our TPRs
	// releaseList := deployv1.ReleaseList{}
	// err = tprclient.Get().Resource(deployv1.ReleaseResourcePath).Do().Into(&releaseList)
	// if err != nil {
	// 	panic(err)
	// }
	// log.Printf("LIST: %#v\n", releaseList)
}
