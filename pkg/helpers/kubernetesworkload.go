package helpers

import (
	"fmt"
	"strings"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/jinzhu/copier"
)

// KubernetesWorkload represents a "base struct" for Kubernetes workloads
type KubernetesWorkload struct {
	metav1.ObjectMeta

	metav1.TypeMeta

	// FriendlyName is the name used to reference the resource in the CLI, ex: deployment/myapp
	FriendlyName string

	// The label selector used to match pods
	PodSelector *metav1.LabelSelector

	// If the resource has a pod template, this should be set
	PodTemplateSpec *corev1.PodTemplateSpec
}

// KubernetesWorkloadReturn is used to fix the issue with channels not supporting pair return values
type KubernetesWorkloadReturn struct {
	Resource *KubernetesWorkload
	Err      error
}

// GetKubernetesWorkload creates a KubernetesWorkload from a StatefulSet
func GetKubernetesWorkload(resource *appsv1.StatefulSet) KubernetesWorkloadReturn {
	copy := KubernetesWorkload{}
	err := copier.Copy(&copy, resource)

	// TypeMeta fields don't seem to always be populated
	copy.Kind = "StatefulSet"
	copy.APIVersion = "apps/v1"

	copy.FriendlyName = fmt.Sprintf("%s/%s", strings.ToLower(copy.Kind), copy.Name)

	copy.PodSelector = resource.Spec.Selector

	copy.PodTemplateSpec = &resource.Spec.Template

	return GetKubernetesWorkloadReturn(&copy, err)
}

// GetKubernetesWorkloadReturn returns a KubernetesWorkloadReturn
func GetKubernetesWorkloadReturn(resource *KubernetesWorkload, err error) KubernetesWorkloadReturn {
	return KubernetesWorkloadReturn{resource, err}
}
