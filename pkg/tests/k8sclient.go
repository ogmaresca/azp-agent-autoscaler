package tests

import (
	"fmt"
	"strings"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/ggmaresca/azp-agent-autoscaler/pkg/args"
	"github.com/ggmaresca/azp-agent-autoscaler/pkg/kubernetes"
)

type mockK8sClient struct {
	Counts    *mockK8sClientCounts
	HPAExists bool
}

// Make this a pointer to allow stateful changes
type mockK8sClientCounts struct {
	NumPods int32
}

// GetWorkload retrieves a Workload with no errors
func (c mockK8sClient) GetWorkloadNoError(args args.KubernetesArgs) *kubernetes.Workload {
	return &kubernetes.Workload{
		ObjectMeta: metav1.ObjectMeta{
			Name:      args.Name,
			Namespace: args.Namespace,
		},
		TypeMeta: metav1.TypeMeta{
			Kind: args.Type,
		},
		FriendlyName: fmt.Sprintf("%s/%s", strings.ToLower(args.Type), args.Name),
		PodSelector: &metav1.LabelSelector{
			MatchLabels: map[string]string{
				"release": args.Name,
			},
		},
		PodTemplateSpec: &corev1.PodTemplateSpec{
			Spec: corev1.PodSpec{},
		},
	}
}

// GetWorkload retrieves a Workload
func (c mockK8sClient) GetWorkload(args args.KubernetesArgs) (*kubernetes.Workload, error) {
	return c.GetWorkloadNoError(args), nil
}

// VerifyNoHorizontalPodAutoscaler returns an error if the given resource has a HorizontalPodAutoscaler
func (c mockK8sClient) VerifyNoHorizontalPodAutoscaler(args args.KubernetesArgs) error {
	if c.HPAExists {
		return fmt.Errorf("Error: %s cannot have a HorizontalPodAutoscaler attached for azp-agent-autoscaler to work", args.FriendlyName())
	}

	return nil
}

// Scale scales a given Kubernetes resource
func (c mockK8sClient) Scale(resource *kubernetes.Workload, replicas int32) error {
	c.Counts.NumPods = replicas
	return nil
}

// GetEnvValue gets an environment variable value from a pod
func (c mockK8sClient) GetEnvValue(podSpec corev1.PodSpec, namespace string, envName string) (string, error) {
	env := kubernetes.GetEnvVar(podSpec, envName)
	if env == nil {
		return "", fmt.Errorf("Could not retrieve environment variable %s", envName)
	}

	if env.Value != "" {
		return env.Value, nil
	}
	return "", fmt.Errorf("Error getting value for environment variable %s", env.Name)
}

// GetPods gets all pods attached to some workload
func (c mockK8sClient) GetPods(workload *kubernetes.Workload) ([]corev1.Pod, error) {
	var pods []corev1.Pod
	for i := int32(0); i < c.Counts.NumPods; i++ {
		pods = append(pods, corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      fmt.Sprintf("%s-%d", workload.Name, i),
				Namespace: workload.Namespace,
			},
			TypeMeta: metav1.TypeMeta{
				Kind: "Pod",
			},
			Spec: workload.PodTemplateSpec.Spec,
			Status: corev1.PodStatus{
				Phase: corev1.PodRunning,
			},
		})
	}
	return pods, nil
}
