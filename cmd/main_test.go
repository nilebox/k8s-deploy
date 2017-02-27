// +build integration

package main

import (
	"net"
	"os"
	"testing"

	"k8s.io/client-go/rest"

	"github.com/nilebox/k8s-deploy/pkg/client"

	deployv1 "github.com/nilebox/k8s-deploy/pkg/apis/v1"
)

func TestCanaryDeployment(t *testing.T) {
	config := configFromEnv(t)

	client, _, err := client.NewClient(config)

	deployment := deployv1.Deployment{
		Spec: deployv1.DeploymentSpec{
			Replicas: 3,
		},
	}
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
