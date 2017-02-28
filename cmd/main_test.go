// +build integration

package main

import (
	"fmt"
	"net"
	"os"
	"testing"

	deployv1 "github.com/nilebox/k8s-deploy/pkg/apis/v1"
	"github.com/nilebox/k8s-deploy/pkg/client"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api"
	"k8s.io/client-go/pkg/api/errors"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/pkg/apis/extensions/v1beta1"
	"k8s.io/client-go/rest"
)

func TestCanaryDeployment(t *testing.T) {
	config := configFromEnv(t)

	clientset, err := kubernetes.NewForConfig(config)

	// initialize third party resource if it does not exist
	tpr, err := clientset.ExtensionsV1beta1().ThirdPartyResources().Get(deployv1.DeploymentResourceName)
	if err != nil {
		if errors.IsNotFound(err) {
			fmt.Printf("NOT FOUND: deployments TPR\n")

			tpr := &v1beta1.ThirdPartyResource{
				ObjectMeta: v1.ObjectMeta{
					Name: deployv1.DeploymentResourceName,
				},
				Versions: []v1beta1.APIVersion{
					{Name: deployv1.DeploymentResourceVersion},
				},
				Description: deployv1.DeploymentResourceDescription,
			}

			result, err := clientset.ExtensionsV1beta1().ThirdPartyResources().Create(tpr)
			if err != nil {
				panic(err)
			}
			fmt.Printf("CREATED: %#v\nFROM: %#v\n", result, tpr)
		} else {
			panic(err)
		}
	} else {
		fmt.Printf("SKIPPING: already exists %#v\n", tpr)
	}

	tprclient, _, err := client.NewClient(config)

	if err != nil {
		panic(err)
	}

	var deployment deployv1.Deployment

	err = tprclient.Get().
		Resource("deployments").
		Namespace(api.NamespaceDefault).
		Name("deploy1").
		Do().Into(&deployment)

	if err != nil {
		if errors.IsNotFound(err) {
			fmt.Printf("NOT FOUND: deployments TPR instance\n")

			// Create an instance of our TPR
			deployment := &deployv1.Deployment{
				ObjectMeta: api.ObjectMeta{
					Name: "deployment1",
				},
				Spec: deployv1.DeploymentSpec{
					Replicas: 3,
				},
			}

			var result deployv1.Deployment
			err = tprclient.Post().
				Resource("deployments").
				Namespace(api.NamespaceDefault).
				Body(deployment).
				Do().Into(&result)

			if err != nil {
				panic(err)
			}
			fmt.Printf("CREATED: %#v\n", result)
		} else {
			panic(err)
		}
	} else {
		fmt.Printf("GET: %#v\n", deployment)
	}

	// Fetch a list of our TPRs
	deploymentList := deployv1.DeploymentList{}
	err = tprclient.Get().Resource("deployments").Do().Into(&deploymentList)
	if err != nil {
		panic(err)
	}
	fmt.Printf("LIST: %#v\n", deploymentList)
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
