package strategy

import (
	"log"
	"time"

	deployv1 "github.com/nilebox/k8s-deploy/pkg/apis/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	apiv1 "k8s.io/client-go/pkg/api/v1"
	v1beta1 "k8s.io/client-go/pkg/apis/extensions/v1beta1"
)

type Canary struct {
	Clientset kubernetes.Interface
}

func (c *Canary) Run(release *deployv1.Release) error {
	trackLabel := "track"
	canaryTrack := "canary"
	stableTrack := "stable"

	// First ensure that canary deployment object exists
	canaryDeployment := &v1beta1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: release.Metadata.Namespace,
			Name:      release.Metadata.Name + "-canary",
		},
		Spec: v1beta1.DeploymentSpec{
			Replicas: release.Spec.Replicas,
			Selector: c.selectorWithLabel(release.Spec.Selector, trackLabel, canaryTrack),
			Template: c.podTemplateWithLabel(release.Spec.Template, trackLabel, canaryTrack),
		},
	}
	err := c.ensureDeploymentExists(canaryDeployment)
	if err != nil {
		log.Printf("Failed to create/update canary deployment: %v", err)
		return err
	}

	// TODO: healthchecks
	log.Printf("Emulate waiting for health check")
	time.Sleep(15 * time.Second) // Wait until the app starts and creates the Release TPR

	// Check if stable deployment object exists too
	stableDeployment := &v1beta1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: release.Metadata.Namespace,
			Name:      release.Metadata.Name + "-stable",
		},
		Spec: v1beta1.DeploymentSpec{
			Replicas: release.Spec.Replicas,
			Selector: c.selectorWithLabel(release.Spec.Selector, trackLabel, stableTrack),
			Template: c.podTemplateWithLabel(release.Spec.Template, trackLabel, stableTrack),
		},
	}
	err = c.ensureDeploymentExists(stableDeployment)
	if err != nil {
		log.Printf("Failed to create/update stable deployment: %v", err)
		return err
	}

	return err
}

func (c *Canary) selectorWithLabel(selector *metav1.LabelSelector, labelName string, labelValue string) *metav1.LabelSelector {
	return &metav1.LabelSelector{
		MatchLabels: c.copyMapWithLabel(selector.MatchLabels, labelName, labelValue),
	}
}

func (c *Canary) podTemplateWithLabel(template apiv1.PodTemplateSpec, labelName string, labelValue string) apiv1.PodTemplateSpec {
	return apiv1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Labels: c.copyMapWithLabel(template.ObjectMeta.Labels, labelName, labelValue),
		},
		Spec: template.Spec,
	}
}

func (c *Canary) copyMapWithLabel(originalMap map[string]string, labelName string, labelValue string) map[string]string {
	newMap := make(map[string]string)
	for k, v := range originalMap {
		newMap[k] = v
	}
	newMap[labelName] = labelValue
	return newMap
}

func (c *Canary) ensureDeploymentExists(deployment *v1beta1.Deployment) error {
	// initialize third party resource if it does not exist
	deployments := c.Clientset.ExtensionsV1beta1().Deployments(deployment.ObjectMeta.Namespace)
	existing, err := deployments.Get(deployment.ObjectMeta.Name, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			log.Printf("NOT FOUND: deployment %s", deployment.ObjectMeta.Name)
			result, err := deployments.Create(deployment)
			if err != nil {
				return err
			}
			log.Printf("CREATED: %s", result.ObjectMeta.SelfLink)
		} else {
			return err
		}
	} else {
		// TODO update existing deployment object
		log.Printf("SKIPPING: already exists %s", existing.ObjectMeta.SelfLink)
	}
	return nil
}
