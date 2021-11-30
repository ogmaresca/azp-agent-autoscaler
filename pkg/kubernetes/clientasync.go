package kubernetes

import (
	"github.com/ogmaresca/azp-agent-autoscaler/pkg/args"

	corev1 "k8s.io/api/core/v1"
)

// WorkloadReturn is used to fix the issue with channels not supporting pair return values
type WorkloadReturn struct {
	Resource *Workload
	Err      error
}

// ClientAsync is a wrapper around the client-go package for Kubernetes
type ClientAsync interface {
	Sync() Client

	GetWorkloadAsync(channel chan<- WorkloadReturn, args args.KubernetesArgs)
	VerifyNoHorizontalPodAutoscalerAsync(channel chan<- error, args args.KubernetesArgs)
	ScaleAsync(channel chan<- error, resource *Workload, replicas int32)
	GetEnvValueAsync(channel chan<- EnvValueReturn, podSpec corev1.PodSpec, namespace string, envName string)
	GetPodsAsync(channel chan<- Pods, workload *Workload)
}

// ClientAsyncImpl is the interface implementation of ClientAsync
type ClientAsyncImpl struct {
	syncClient Client
}

// MakeClient returns a ClientAsync
func MakeClient() (ClientAsync, error) {
	syncClient, err := makeClient()
	if err == nil {
		return ClientAsyncImpl{syncClient}, nil
	}
	return nil, err
}

// MakeFromClient returns a ClientAsync from the given sync client
func MakeFromClient(syncClient Client) ClientAsync {
	return ClientAsyncImpl{syncClient}
}

// Sync returns the synchronous client
func (c ClientAsyncImpl) Sync() Client {
	return c.syncClient
}

// GetWorkloadAsync retrieves a Workload
func (c ClientAsyncImpl) GetWorkloadAsync(channel chan<- WorkloadReturn, args args.KubernetesArgs) {
	workload, err := c.syncClient.GetWorkload(args)
	channel <- WorkloadReturn{workload, err}
}

// VerifyNoHorizontalPodAutoscalerAsync returns an error if the given resource has a HorizontalPodAutoscaler
func (c ClientAsyncImpl) VerifyNoHorizontalPodAutoscalerAsync(channel chan<- error, args args.KubernetesArgs) {
	channel <- c.syncClient.VerifyNoHorizontalPodAutoscaler(args)
}

// ScaleAsync scales a given Kubernetes resource
func (c ClientAsyncImpl) ScaleAsync(channel chan<- error, resource *Workload, replicas int32) {
	channel <- c.syncClient.Scale(resource, replicas)
}

// EnvValueReturn is a wrapper around string to allow returning multiple values in a channel
type EnvValueReturn struct {
	Value string
	Err   error
}

// GetEnvValueAsync gets an environment variable value from a pod
func (c ClientAsyncImpl) GetEnvValueAsync(channel chan<- EnvValueReturn, podSpec corev1.PodSpec, namespace string, envName string) {
	value, err := c.syncClient.GetEnvValue(podSpec, namespace, envName)
	channel <- EnvValueReturn{value, err}
}

// Pods is a wrapper around []corev1.Pod to allow returning multiple values in a channel
type Pods struct {
	Pods []corev1.Pod
	Err  error
}

// GetPodsAsync gets all pods attached to some workload
func (c ClientAsyncImpl) GetPodsAsync(channel chan<- Pods, workload *Workload) {
	value, err := c.syncClient.GetPods(workload)
	channel <- Pods{value, err}
}
