package kubernetes

import (
	"encoding/base64"
	"fmt"
	"os"
	"strings"

	"github.com/ogmaresca/azp-agent-autoscaler/pkg/args"
	autoscalingv1 "k8s.io/api/autoscaling/v1"
	corev1 "k8s.io/api/core/v1"
	apimachinery "k8s.io/apimachinery/pkg/apis/meta/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8s "k8s.io/client-go/kubernetes"
	k8srest "k8s.io/client-go/rest"
	k8sclientcmd "k8s.io/client-go/tools/clientcmd"
)

// Client is a wrapper around the client-go package for Kubernetes
type Client interface {
	GetWorkload(args args.KubernetesArgs) (*Workload, error)
	VerifyNoHorizontalPodAutoscaler(args args.KubernetesArgs) error
	Scale(resource *Workload, replicas int32) error
	GetEnvValue(podSpec corev1.PodSpec, namespace string, envName string) (string, error)
	GetPods(workload *Workload) ([]corev1.Pod, error)
}

// ClientImpl is the interface implementation of Client
type ClientImpl struct {
	client *k8s.Clientset
}

// makeClient returns a Client
func makeClient() (Client, error) {
	k8sConfig, err := k8srest.InClusterConfig()
	if err != nil {
		kubeconfigEnv := os.Getenv("KUBECONFIG")
		k8sConfig, err = k8sclientcmd.BuildConfigFromFlags("", kubeconfigEnv)
		if err != nil {
			home := os.Getenv("HOME")
			if home == "" {
				home = os.Getenv("USERPROFILE") // windows
			}
			k8sConfig, err = k8sclientcmd.BuildConfigFromFlags("", fmt.Sprintf("%s/.kube/config", home))
			if err != nil {
				return nil, fmt.Errorf("Error initializing Kubernetes config: %s", err.Error())
			}
		}
	}

	clientset, err := k8s.NewForConfig(k8sConfig)
	if err != nil {
		return nil, err
	}
	return ClientImpl{clientset}, nil
}

// GetWorkload retrieves a Workload
func (c ClientImpl) GetWorkload(args args.KubernetesArgs) (*Workload, error) {
	if strings.EqualFold(args.Type, "StatefulSet") {
		return c.getStatefulSet(args.Namespace, args.Name)
	} else {
		return nil, fmt.Errorf("Resource kind %s is not implemented", args.Type)
	}
}

func (c ClientImpl) getStatefulSet(namespace string, name string) (*Workload, error) {
	statefulSet, err := c.client.AppsV1().StatefulSets(namespace).Get(name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	} else if statefulSet == nil {
		return nil, fmt.Errorf("Could not find statefulset/%s in namespace %s", name, namespace)
	} else {
		return GetWorkload(statefulSet)
	}
}

// VerifyNoHorizontalPodAutoscaler returns an error if the given resource has a HorizontalPodAutoscaler
func (c ClientImpl) VerifyNoHorizontalPodAutoscaler(args args.KubernetesArgs) error {
	hpas, err := c.client.AutoscalingV1().HorizontalPodAutoscalers(args.Namespace).List(metav1.ListOptions{})
	if err != nil {
		return err
	}
	for _, hpa := range hpas.Items {
		if strings.EqualFold(hpa.Spec.ScaleTargetRef.Kind, args.Type) && hpa.Spec.ScaleTargetRef.Name == args.Name {
			return fmt.Errorf("Error: %s cannot have a HorizontalPodAutoscaler attached for azp-agent-autoscaler to work", args.FriendlyName())
		}
	}

	return nil
}

// Scale scales a given Kubernetes resource
func (c ClientImpl) Scale(resource *Workload, replicas int32) error {
	var getScaleFunc func() (*autoscalingv1.Scale, error)
	var doScaleFunc func(scale *autoscalingv1.Scale) error
	if strings.EqualFold(resource.Kind, "StatefulSet") {
		statefulsets := c.client.AppsV1().StatefulSets(resource.Namespace)
		getScaleFunc = func() (*autoscalingv1.Scale, error) {
			return statefulsets.GetScale(resource.Name, metav1.GetOptions{})
		}
		doScaleFunc = func(scale *autoscalingv1.Scale) error {
			scale, err := statefulsets.UpdateScale(resource.Name, scale)
			return err
		}
	} else {
		return fmt.Errorf("Resource kind %s is not implemented", resource.Kind)
	}

	scale, err := getScaleFunc()
	if err != nil {
		return err
	}
	if scale.Spec.Replicas == replicas {
		return nil
	}
	scale.Spec.Replicas = replicas
	return doScaleFunc(scale)
}

// GetEnvValue gets an environment variable value from a pod
func (c ClientImpl) GetEnvValue(podSpec corev1.PodSpec, namespace string, envName string) (string, error) {
	env := GetEnvVar(podSpec, envName)
	if env == nil {
		return "", fmt.Errorf("Could not retrieve environment variable %s", envName)
	}

	if env.Value != "" {
		return env.Value, nil
	} else if env.ValueFrom != nil {
		if env.ValueFrom.FieldRef != nil {
			return "", fmt.Errorf("Error getting value for environment variable %s: fieldRef is not supported", env.Name)
		} else if env.ValueFrom.ResourceFieldRef != nil {
			return "", fmt.Errorf("Error getting value for environment variable %s: resourceFieldRef is not supported", env.Name)
		} else if env.ValueFrom.ConfigMapKeyRef != nil {
			configmap, err := c.client.CoreV1().ConfigMaps(namespace).Get(env.ValueFrom.ConfigMapKeyRef.Name, metav1.GetOptions{})
			if err != nil {
				return "", fmt.Errorf("Error getting value from configmap %s for environment variable %s: %s", env.ValueFrom.ConfigMapKeyRef.Name, envName, err.Error())
			}
			value, exists := configmap.Data[env.ValueFrom.ConfigMapKeyRef.Key]
			if !exists {
				return "", fmt.Errorf("Error getting value from configmap %s for environment variable %s: key %s does not exist", env.ValueFrom.ConfigMapKeyRef.Name, envName, env.ValueFrom.ConfigMapKeyRef.Key)
			}
			return value, nil
		} else if env.ValueFrom.SecretKeyRef != nil {
			secret, err := c.client.CoreV1().Secrets(namespace).Get(env.ValueFrom.SecretKeyRef.Name, metav1.GetOptions{})
			if err != nil {
				return "", fmt.Errorf("Error getting value from secret %s for environment variable %s: %s", env.ValueFrom.SecretKeyRef.Name, envName, err.Error())
			}
			value, exists := secret.Data[env.ValueFrom.SecretKeyRef.Key]
			if !exists {
				return "", fmt.Errorf("Error getting value from secret %s for environment variable %s: key %s does not exist", env.ValueFrom.SecretKeyRef.Name, envName, env.ValueFrom.SecretKeyRef.Key)
			}
			decodedValue, err := base64.StdEncoding.DecodeString(string(value))
			if err != nil {
				return "", fmt.Errorf("Error getting value from secret %s for environment variable %s: could not decode value for %s", env.ValueFrom.SecretKeyRef.Name, envName, env.ValueFrom.SecretKeyRef.Key)
			}
			return string(decodedValue), nil
		}
	}
	return "", fmt.Errorf("Error getting value for environment variable %s", env.Name)
}

// GetPods gets all pods attached to some workload
func (c ClientImpl) GetPods(workload *Workload) ([]corev1.Pod, error) {
	listOptions := metav1.ListOptions{
		LabelSelector: apimachinery.FormatLabelSelector(workload.PodSelector),
	}
	pods, err := c.client.CoreV1().Pods(workload.Namespace).List(listOptions)
	if err != nil {
		return nil, err
	}
	return pods.Items, nil
}
