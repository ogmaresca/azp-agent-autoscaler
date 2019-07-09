package kubernetes

import corev1 "k8s.io/api/core/v1"

// GetEnvVar find the first EnvVar with the provided environment name from a PodSpec
func GetEnvVar(podSpec corev1.PodSpec, envName string) *corev1.EnvVar {
	for _, container := range podSpec.Containers {
		for _, containerEnv := range container.Env {
			if containerEnv.Name == envName {
				return &containerEnv
			}
		}
	}
	return nil
}
