package helpers

import (
	"fmt"
	"strings"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/jinzhu/copier"
)

// KubernetesResource represents a "base struct" for Kubernetes resources
type KubernetesResource struct {
	metav1.ObjectMeta

	metav1.TypeMeta

	// FriendlyName is the name used to reference the resource in the CLI, ex: deployment/myapp
	FriendlyName string

	// If the resource has a pod template, this should be set
	PodTemplateSpec *corev1.PodTemplateSpec
}

// KubernetesResourceReturn is used to fix the issue with channels not supporting pair return values
type KubernetesResourceReturn struct {
	Resource *KubernetesResource
	Err      error
}

// GetKubernetesResource creates a KubernetesResource from a StatefulSet
func GetKubernetesResource(resource *appsv1.StatefulSet) KubernetesResourceReturn {
	copy := KubernetesResource{}
	err := copier.Copy(&copy, resource)

	// TypeMeta fields don't seem to always be populated
	copy.Kind = "StatefulSet"
	copy.APIVersion = "apps/v1"

	copy.FriendlyName = fmt.Sprintf("%s/%s", strings.ToLower(copy.Kind), copy.Name)

	copy.PodTemplateSpec = &resource.Spec.Template

	return GetKubernetesResourceReturn(&copy, err)
}

// GetKubernetesResourceReturn returns a KubernetesResourceReturn
func GetKubernetesResourceReturn(resource *KubernetesResource, err error) KubernetesResourceReturn {
	return KubernetesResourceReturn{resource, err}
}
