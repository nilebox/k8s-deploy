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

func TestCanaryCompute(t *testing.T) {
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

	time.Sleep(5 * time.Second) // Wait until the app starts and creates the Compute TPR

	tprclient, _, err := client.NewClient(config)

	if err != nil {
		panic(err)
	}

	var compute deployv1.Compute

	err = tprclient.Get().
		Resource(deployv1.ComputeResourcePath).
		Namespace(api.NamespaceDefault).
		Name("compute2").
		Do().Into(&compute)

	if err != nil {
		if errors.IsNotFound(err) {
			log.Printf("NOT FOUND: compute2 instance\n")

			replicas := int32(3)
			// Create an instance of our TPR
			compute := &deployv1.Compute{
				TypeMeta: metav1.TypeMeta{
					APIVersion: deployv1.ComputeResourceGroupVersion,
					Kind:       deployv1.ComputeResourceKind,
				},
				Metadata: apiv1.ObjectMeta{
					Name: "compute2",
				},
				Spec: deployv1.ComputeSpec{
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
								{
									Name:  "tea",
									Image: "nginxdemos/hello",
									Ports: []apiv1.ContainerPort{
										{
											ContainerPort: 80,
										},
									},
								},
							},
						},
					},
				},
			}

			var result deployv1.Compute
			err = tprclient.Post().
				Resource(deployv1.ComputeResourcePath).
				Namespace(api.NamespaceDefault).
				Body(compute).
				Do().Into(&result)

			if err != nil {
				panic(err)
			}
			log.Printf("CREATED: %#v\n", result)
		} else {
			panic(err)
		}
	} else {
		log.Printf("GET: %#v\n", compute)
	}

	// // Fetch a list of our TPRs
	// computeList := deployv1.ComputeList{}
	// err = tprclient.Get().Resource(deployv1.ComputeResourcePath).Do().Into(&computeList)
	// if err != nil {
	// 	panic(err)
	// }
	// log.Printf("LIST: %#v\n", computeList)
}
