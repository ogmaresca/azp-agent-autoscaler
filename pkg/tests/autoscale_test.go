package tests

import (
	"fmt"
	"testing"
	"time"

	"github.com/ggmaresca/azp-agent-autoscaler/pkg/args"
	"github.com/ggmaresca/azp-agent-autoscaler/pkg/kubernetes"
	"github.com/ggmaresca/azp-agent-autoscaler/pkg/math"
	"github.com/ggmaresca/azp-agent-autoscaler/pkg/scaling"
)

var (
	agentPoolID = 2
)

func TestAutoscale(t *testing.T) {
	failed := false
	for numActiveJobs := int32(0); numActiveJobs < 10; numActiveJobs = numActiveJobs + 1 {
		for numQueuedJobs := int32(0); numQueuedJobs < 10; numQueuedJobs = numQueuedJobs + 1 {
			for numFreeAgents := int32(0); numFreeAgents < 20; numFreeAgents = numFreeAgents + 1 {
				for min := int32(1); min < 13; min = min + 3 {
					for max := min + 1; max < min+20; max = max + 3 {
						if failed {
							break
						}
						t.Run(fmt.Sprintf("%d active & %d queued jobs, %d agents, min %d, max %d", numActiveJobs, numQueuedJobs, numFreeAgents, min, max), func(t *testing.T) {
							azdClient := mockAZDClient{
								NumPools:         5,
								ErrorListPools:   false,
								NumFreeAgents:    numFreeAgents,
								NumRunningAgents: numActiveJobs,
								ErrorAgents:      false,
								NumQueuedJobs:    numQueuedJobs,
								ErrorJobs:        false,
							}

							args := args.Args{
								Min:  min,
								Max:  max,
								Rate: 10 * time.Second,
								ScaleDown: args.ScaleDownArgs{
									Delay: 0 * time.Nanosecond,
									Max:   int32(100),
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

							originalPodCount := numActiveJobs + numFreeAgents
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
							expectedPodCount := math.MaxInt32(numActiveJobs, math.MinInt32(numActiveJobs+numQueuedJobs+min, max))
							if k8sClient.Counts.NumPods != expectedPodCount {
								failed = true
								t.Fatalf("Expected %d pods (from %d), but got %d", expectedPodCount, originalPodCount, k8sClient.Counts.NumPods)
							}
						})
					}
				}
			}
		}
	}
}
