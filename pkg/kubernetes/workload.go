package kubernetes

import (
	"fmt"
	"strings"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/jinzhu/copier"
)

// Workload represents a "base struct" for Kubernetes workloads
type Workload struct {
	metav1.ObjectMeta

	metav1.TypeMeta

	// FriendlyName is the name used to reference the resource in the CLI, ex: deployment/myapp
	FriendlyName string

	// The label selector used to match pods
	PodSelector *metav1.LabelSelector

	// If the resource has a pod template, this should be set
	PodTemplateSpec *corev1.PodTemplateSpec
}

// GetWorkload creates a KubernetesWorkload from a StatefulSet
func GetWorkload(resource *appsv1.StatefulSet) (*Workload, error) {
	copy := Workload{}
	err := copier.Copy(&copy, resource)

	// TypeMeta fields don't seem to always be populated
	copy.Kind = "StatefulSet"
	copy.APIVersion = "apps/v1"

	copy.FriendlyName = fmt.Sprintf("%s/%s", strings.ToLower(copy.Kind), copy.Name)

	copy.PodSelector = resource.Spec.Selector

	copy.PodTemplateSpec = &resource.Spec.Template

	return &copy, err
}
