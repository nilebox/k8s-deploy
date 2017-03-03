package strategy

import (
	"log"

	deployv1 "github.com/nilebox/k8s-deploy/pkg/apis/v1"
	"k8s.io/client-go/kubernetes"
	apierrors "k8s.io/client-go/pkg/api/errors"
	apiv1 "k8s.io/client-go/pkg/api/v1"
	v1beta1 "k8s.io/client-go/pkg/apis/extensions/v1beta1"
)

type Canary struct {
	Clientset kubernetes.Interface
}

func (c *Canary) Run(release *deployv1.Release) error {
	// First ensure that canary deployment object exists
	canaryDeployment := &v1beta1.Deployment{
		ObjectMeta: apiv1.ObjectMeta{
			Namespace: release.ObjectMeta.Namespace,
			Name:      release.ObjectMeta.Name + "-canary",
		},
		Spec: v1beta1.DeploymentSpec{
			Replicas: release.Spec.Replicas,
			Selector: release.Spec.Selector,
			Template: release.Spec.Template,
		},
	}
	err := c.ensureDeploymentExists(canaryDeployment)
	if err != nil {
		log.Printf("Failed to create/update canary deployment: %v", err)
		return err
	}
	// TODO: healthchecks

	// Check if stable deployment object exists too
	stableDeployment := &v1beta1.Deployment{
		ObjectMeta: apiv1.ObjectMeta{
			Namespace: release.ObjectMeta.Namespace,
			Name:      release.ObjectMeta.Name + "-stable",
		},
		Spec: v1beta1.DeploymentSpec{
			Replicas: release.Spec.Replicas,
			Selector: release.Spec.Selector,
			Template: release.Spec.Template,
		},
	}
	err = c.ensureDeploymentExists(stableDeployment)
	if err != nil {
		log.Printf("Failed to create/update stable deployment: %v", err)
		return err
	}

	return err
}

func (c *Canary) ensureDeploymentExists(deployment *v1beta1.Deployment) error {
	// initialize third party resource if it does not exist
	deployments := c.Clientset.ExtensionsV1beta1().Deployments(deployment.ObjectMeta.Namespace)
	tpr, err := deployments.Get(deployment.ObjectMeta.Name)
	if err != nil {
		if apierrors.IsNotFound(err) {
			log.Printf("NOT FOUND: deployment %s", deployment.ObjectMeta.Name)
			result, err := deployments.Create(deployment)
			if err != nil {
				return err
			}
			log.Printf("CREATED: %#v\nFROM: %#v\n", result, tpr)
		} else {
			return err
		}
	} else {
		// TODO update existing deployment object
		log.Printf("SKIPPING: already exists %#v\n", tpr)
	}
	return nil
}