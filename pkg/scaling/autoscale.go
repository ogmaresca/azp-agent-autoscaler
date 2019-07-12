package scaling

import (
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"

	"github.com/ggmaresca/azp-agent-autoscaler/pkg/args"
	"github.com/ggmaresca/azp-agent-autoscaler/pkg/azuredevops"
	"github.com/ggmaresca/azp-agent-autoscaler/pkg/collections"
	"github.com/ggmaresca/azp-agent-autoscaler/pkg/kubernetes"
	"github.com/ggmaresca/azp-agent-autoscaler/pkg/logging"
	"github.com/ggmaresca/azp-agent-autoscaler/pkg/math"
)

var (
	lastScaleDown = time.Date(1970, time.January, 1, 0, 0, 0, 0, time.UTC)
)

// Autoscale the agent deployment
func Autoscale(azdClient azuredevops.ClientAsync, agentPoolID int, k8sClient kubernetes.ClientAsync, deployment *kubernetes.Workload, args args.Args) error {
	agentsChan := make(chan azuredevops.PoolAgentsResponse)
	jobsChan := make(chan azuredevops.JobRequestsResponse)
	podsChan := make(chan kubernetes.Pods)

	// Get all active agents
	go azdClient.ListPoolAgentsAsync(agentsChan, agentPoolID)
	// Get all queued jobs
	go azdClient.ListJobRequestsAsync(jobsChan, agentPoolID)
	// Get all pods
	go k8sClient.GetPodsAsync(podsChan, deployment)

	agents := <-agentsChan
	if agents.Err != nil {
		return agents.Err
	}
	jobs := <-jobsChan
	if jobs.Err != nil {
		return jobs.Err
	}
	pods := <-podsChan
	if pods.Err != nil {
		return pods.Err
	}

	// Get all pod names and statuses
	podNames := make(collections.StringSet)
	numPods := int32(len(pods.Pods))
	numRunningPods, numPendingPods, numUnschedulablePods := int32(0), int32(0), int32(0)
	for _, pod := range pods.Pods {
		podNames.Add(pod.Name)
		if pod.Status.Phase == corev1.PodRunning {
			allContainersRunning := true
			for _, containerStatus := range pod.Status.ContainerStatuses {
				if containerStatus.State.Running == nil || containerStatus.State.Terminated != nil {
					allContainersRunning = false
					break
				}
			}
			if allContainersRunning {
				numRunningPods = numRunningPods + 1
			} else {
				numPendingPods = numPendingPods + 1
			}
		} else if pod.Status.Phase == corev1.PodPending {
			numPendingPods = numPendingPods + 1
			for _, podCondition := range pod.Status.Conditions {
				if podCondition.Type == corev1.PodScheduled && podCondition.Status == corev1.ConditionFalse && podCondition.Reason == corev1.PodReasonUnschedulable {
					numUnschedulablePods = numUnschedulablePods + 1
				}
			}
		}
	}
	numFailedPods := numPods - numRunningPods - numPendingPods

	logging.Logger.Tracef("%d pods (%d running, %d pending, %d failed)", numPods, numRunningPods, numPendingPods, numFailedPods)

	// Get number of active agents
	activeAgentNames := getActiveAgentNames(agents.Agents, podNames)
	numActiveAgents := int32(len(activeAgentNames))

	// Determine the number of jobs that are queued
	numQueuedJobs := getNumQueuedJobs(jobs.Jobs, activeAgentNames)

	logging.Logger.Debugf("Found %d active agents out of %d agents in the cluster. There are %d queued jobs.", numActiveAgents, numPods, numQueuedJobs)

	if numRunningPods != numPods {
		if !(numUnschedulablePods == numPendingPods && numFailedPods == 0) {
			logging.Logger.Infof("Not scaling - there are %d pending pods and %d failed pods.", numPendingPods, numFailedPods)
			return nil
		}
	}

	// Determine delta for how much to scale by
	scale := int32(0)
	if numActiveAgents+numQueuedJobs+args.Min > numPods {
		// Scale up
		scale = numActiveAgents + numQueuedJobs + args.Min - numPods
	} else if numActiveAgents+args.Min+numQueuedJobs < numPods {
		// Scale down
		scale = -numPods + numActiveAgents + args.Min + numQueuedJobs
	}

	if scale > 0 && numUnschedulablePods > 0 {
		logging.Logger.Infof("Not scaling up - there are %d unschedulable pods.", numUnschedulablePods)
		return nil
	}

	// Apply scaling limits and scale down limits
	podsToScaleTo := numPods
	if scale > 0 {
		podsToScaleTo = math.MaxInt32(numActiveAgents, math.MinInt32(args.Max, numPods+scale))
	} else if scale < 0 {
		podsToScaleTo = math.MaxInt32(numActiveAgents, math.MinInt32(args.Max, math.MaxInt32(args.Min, numPods+scale)))

		now := time.Now()
		if now.Before(lastScaleDown.Add(args.ScaleDown.Delay)) {
			logging.Logger.Debugf("Not scaling down %s from %d to %d pods - cannot scale down until %s", deployment.FriendlyName, numPods, podsToScaleTo, lastScaleDown.Add(args.ScaleDown.Delay).String())
		} else if numPods-podsToScaleTo > args.ScaleDown.Max {
			logging.Logger.Debugf("Capping the scale down from %d to %d pods", podsToScaleTo, numPods-args.ScaleDown.Max)
			podsToScaleTo = numPods - args.ScaleDown.Max
		}
	} else if podsToScaleTo > args.Max {
		if numActiveAgents > args.Max {
			podsToScaleTo = numActiveAgents
			logging.Logger.Warningf("There are %d pods over the max of %d - limiting the scale down to %d active agents", numPods, args.Max, numActiveAgents)
		} else {
			podsToScaleTo = args.Max
			logging.Logger.Warningf("There are %d pods over the max of %d - scaling down to meet the max", numPods, args.Max)
		}
	} else {
		logging.Logger.Tracef("Not scaling %s from %d pods", deployment.FriendlyName, numPods)
		return nil
	}

	if numPods != podsToScaleTo {
		logging.Logger.Infof("Scaling %s from %d to %d pods", deployment.FriendlyName, numPods, podsToScaleTo)
		err := k8sClient.Sync().Scale(deployment, podsToScaleTo)
		if err == nil && scale < 0 {
			lastScaleDown = time.Now()
		}
		return err
	}

	logging.Logger.Debugf("Not scaling from %d pods", numPods)
	return nil
}

func getActiveAgentNames(agents []azuredevops.AgentDetails, podNames collections.StringSet) collections.StringSet {
	activeAgentNames := make(collections.StringSet)
	for _, agent := range agents {
		if strings.EqualFold(agent.Status, "online") {
			podName := agent.SystemCapabilities["HOSTNAME"]
			if podNames.Contains(podName) && agent.AssignedRequest != nil {
				activeAgentNames.Add(agent.Name)
			}
		}
	}
	return activeAgentNames
}

func getNumQueuedJobs(jobs []azuredevops.JobRequest, activeAgentNames collections.StringSet) int32 {
	numQueuedJobs := int32(0)
	for _, job := range jobs {
		if job.IsQueuedOrRunning() && job.ReservedAgent == nil {
			if job.MatchesAllAgentsInPool {
				numQueuedJobs = numQueuedJobs + 1
			} else if len(job.MatchedAgents) > 0 {
				for _, agent := range job.MatchedAgents {
					if activeAgentNames.Contains(agent.Name) {
						numQueuedJobs = numQueuedJobs + 1
						break
					}
				}
			}
		}
	}
	return numQueuedJobs
}
