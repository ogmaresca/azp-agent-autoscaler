package tests

import (
	"fmt"
	"testing"
	"time"

	"github.com/ogmaresca/azp-agent-autoscaler/pkg/args"
	"github.com/ogmaresca/azp-agent-autoscaler/pkg/kubernetes"
	"github.com/ogmaresca/azp-agent-autoscaler/pkg/math"
	"github.com/ogmaresca/azp-agent-autoscaler/pkg/scaling"
)

var (
	agentPoolID = 2
)

func TestAutoscale(t *testing.T) {
	failed := false
	for numActiveJobs := int32(0); numActiveJobs < 10; numActiveJobs = numActiveJobs + 1 {
		for numQueuedJobs := int32(0); numQueuedJobs < 10; numQueuedJobs = numQueuedJobs + 1 {
			for numFreeAgents := int32(0); numFreeAgents < 20; numFreeAgents = numFreeAgents + 1 {
				originalPodCount := numActiveJobs + numFreeAgents

				for min := int32(1); min < 13; min = min + 3 {
					for max := min + 1; max < min+20; max = max + 3 {
						expectedPodCount := math.MaxInt32(numActiveJobs, math.MinInt32(numActiveJobs+numQueuedJobs+min, max))

						for maxScaleDown := int32(1); maxScaleDown < math.MaxInt32(1, originalPodCount-expectedPodCount); maxScaleDown++ {
							if failed {
								break
							}

							testName := fmt.Sprintf("%d_activejobs,%d_queuedjobs,%d_agents,%d_min,%d_max,%d_scaledownmax", numActiveJobs, numQueuedJobs, numFreeAgents, min, max, maxScaleDown)

							expectedPodCountWithScaleDownMax := expectedPodCount
							if expectedPodCount < originalPodCount {
								expectedPodCountWithScaleDownMax = math.MaxInt32(originalPodCount-maxScaleDown, expectedPodCount)
							}

							t.Run(testName, func(t *testing.T) {
								azdClient := mockAZDClient{
									NumPools:         5,
									ErrorListPools:   false,
									NumFreeAgents:    numFreeAgents,
									NumRunningAgents: numActiveJobs,
									ErrorAgents:      false,
									NumQueuedJobs:    numQueuedJobs,
									ErrorJobs:        false,
									FreeAgentsFirst:  false,
								}

								args := args.Args{
									Min:  min,
									Max:  max,
									Rate: 10 * time.Second,
									ScaleDown: args.ScaleDownArgs{
										Delay: 0 * time.Nanosecond,
										Max:   maxScaleDown,
									},
									Kubernetes: args.KubernetesArgs{
										Type:      "StatefulSet",
										Name:      "azp-agent",
										Namespace: "default",
									},
									AZD: args.AzureDevopsArgs{
										Token: "azdtoken",
										URL:   "https://dev.azure.com/organization",
									},
								}

								k8sClient := mockK8sClient{
									Counts: &mockK8sClientCounts{
										NumPods: originalPodCount,
									},
									HPAExists: false,
								}

								err := scaling.Autoscale(azdClient, agentPoolID, kubernetes.MakeFromClient(k8sClient), k8sClient.GetWorkloadNoError(args.Kubernetes), args)
								if err != nil {
									t.Error(err.Error())
								}

								if k8sClient.Counts.NumPods != expectedPodCountWithScaleDownMax {
									failed = true
									t.Fatalf("Expected %d pods (%d without scale-down limits, from %d), but got %d", expectedPodCountWithScaleDownMax, expectedPodCount, originalPodCount, k8sClient.Counts.NumPods)
								}
							})
						}
					}
				}
			}
		}
	}
}

func TestAutoscalePodNames(t *testing.T) {
	t.Run("verify_pod_name_scale_down_limit", func(t *testing.T) {
		azdClient := mockAZDClient{
			NumPools:         5,
			ErrorListPools:   false,
			NumFreeAgents:    10,
			NumRunningAgents: 1,
			ErrorAgents:      false,
			NumQueuedJobs:    0,
			ErrorJobs:        false,
			FreeAgentsFirst:  true,
		}

		args := args.Args{
			Min:  1,
			Max:  100,
			Rate: 10 * time.Second,
			ScaleDown: args.ScaleDownArgs{
				Delay: 0 * time.Nanosecond,
				Max:   100,
			},
			Kubernetes: args.KubernetesArgs{
				Type:      "StatefulSet",
				Name:      "azp-agent",
				Namespace: "default",
			},
			AZD: args.AzureDevopsArgs{
				Token: "azdtoken",
				URL:   "https://dev.azure.com/organization",
			},
		}

		expectedPodCount := azdClient.NumFreeAgents + azdClient.NumRunningAgents

		k8sClient := mockK8sClient{
			Counts: &mockK8sClientCounts{
				NumPods: expectedPodCount,
			},
			HPAExists: false,
		}

		err := scaling.Autoscale(azdClient, agentPoolID, kubernetes.MakeFromClient(k8sClient), k8sClient.GetWorkloadNoError(args.Kubernetes), args)
		if err != nil {
			t.Error(err.Error())
		}

		if k8sClient.Counts.NumPods != expectedPodCount {
			t.Fatalf("Expected %d pods (no scale down), but got %d", expectedPodCount, k8sClient.Counts.NumPods)
		}
	})
}
