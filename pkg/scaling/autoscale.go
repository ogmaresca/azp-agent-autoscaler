package scaling

import (
	"fmt"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"

	"github.com/ggmaresca/azp-agent-autoscaler/pkg/args"
	"github.com/ggmaresca/azp-agent-autoscaler/pkg/azuredevops"
	"github.com/ggmaresca/azp-agent-autoscaler/pkg/collections"
	"github.com/ggmaresca/azp-agent-autoscaler/pkg/kubernetes"
	"github.com/ggmaresca/azp-agent-autoscaler/pkg/logging"
	"github.com/ggmaresca/azp-agent-autoscaler/pkg/math"
)

var (
	lastScaleDown    = time.Date(1970, time.January, 1, 0, 0, 0, 0, time.UTC)
	scaleDownCounter = promauto.NewCounter(prometheus.CounterOpts{
		Name: "azp_agent_autoscaler_scale_down_count",
		Help: "The total number of scale downs",
	})
	scaleUpCounter = promauto.NewCounter(prometheus.CounterOpts{
		Name: "azp_agent_autoscaler_scale_up_count",
		Help: "The total number of scale ups",
	})
	scaleDownLimitedCounter = promauto.NewCounter(prometheus.CounterOpts{
		Name: "azp_agent_autoscaler_scale_down_limited_count",
		Help: "The total number of scale downs prevented due to limits",
	})
	scaleSizeGauge = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "azp_agent_autoscaler_scale_size",
		Help: "The size of the agent scaling",
	})
	totalAgentsGauge = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "azp_agent_autoscaler_total_agents_count",
		Help: "The total number of agents",
	})
	activeAgentsGauge = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "azp_agent_autoscaler_active_agents_count",
		Help: "The number of active agents",
	})
	pendingAgentsGauge = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "azp_agent_autoscaler_pending_agents_count",
		Help: "The number of pending agents",
	})
	failedAgentsGauge = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "azp_agent_autoscaler_failed_agents_count",
		Help: "The number of failed agents",
	})
	queuedPodsGauge = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "azp_agent_autoscaler_queued_pods_count",
		Help: "The number of queued pods",
	})
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
	activeAgentPodNames := getActiveAgentPodNames(agents.Agents, podNames)
	numActiveAgents := int32(len(activeAgentNames))

	// Determine the number of jobs that are queued
	numQueuedJobs := getNumQueuedJobs(jobs.Jobs, activeAgentNames)

	logging.Logger.Debugf("Found %d active agents out of %d agents in the cluster. There are %d queued jobs.", numActiveAgents, numPods, numQueuedJobs)

	// Apply metrics
	totalAgentsGauge.Set(float64(numPods))
	activeAgentsGauge.Set(float64(numActiveAgents))
	pendingAgentsGauge.Set(float64(numPendingPods))
	failedAgentsGauge.Set(float64(numFailedPods))
	queuedPodsGauge.Set(float64(numQueuedJobs))

	if numRunningPods != numPods {
		if !(numUnschedulablePods == numPendingPods && numFailedPods == 0) {
			logging.Logger.Infof("Not scaling - there are %d pending pods and %d failed pods.", numPendingPods, numFailedPods)
			scaleSizeGauge.Set(0)
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

	// Allow scaling down if there are unschedulable pods
	// This way node(s) don't have to be allocated and all of the pods launched before a scale down is allowed
	if scale > 0 && numUnschedulablePods > 0 {
		logging.Logger.Infof("Not scaling up - there are %d unschedulable pods.", numUnschedulablePods)
		scaleSizeGauge.Set(0)
		return nil
	}

	// If there are currently 10 pods and 1 active job, but azp-agent-9 (statefulset pod names are zero-indexed)
	// is currently active, then don't scale down
	if scale < 0 && numActiveAgents > 0 && strings.EqualFold(deployment.Kind, "StatefulSet") {
		maxActivePod := int32(0)
		for i := numPods - 1; i > 0 && maxActivePod == 0; i-- {
			if activeAgentPodNames.Contains(fmt.Sprintf("%s-%d", deployment.Name, i)) {
				maxActivePod = i
				break
			}
		}
		if maxActivePod > 0 {
			scale = math.MaxInt32(0-numPods+1+maxActivePod, scale)
			if scale == 0 {
				logging.Logger.Debugf("Not scaling down - the last agent pod is active")
				scaleSizeGauge.Set(0)
				return nil
			}
		}
	}

	// Apply scaling limits and scale down limits
	podsToScaleTo := numPods
	if scale > 0 {
		// Scale up
		podsToScaleTo = math.MaxInt32(numActiveAgents, math.MinInt32(args.Max, numPods+scale), numPods-args.ScaleDown.Max)
	} else if scale < 0 {
		// Scale down, don't kill active agents
		podsToScaleTo = math.MaxInt32(numActiveAgents, math.MinInt32(args.Max, math.MaxInt32(args.Min, numPods+scale)))
	} else if podsToScaleTo > args.Max {
		// If there happens to be more pods than the max arg
		if numActiveAgents > args.Max {
			podsToScaleTo = numActiveAgents
			logging.Logger.Warningf("There are %d pods over the max of %d - limiting the scale down to %d active agents", numPods, args.Max, numActiveAgents)
		} else {
			podsToScaleTo = math.MaxInt32(args.Max, numPods-args.ScaleDown.Max)
			logging.Logger.Warningf("There are %d pods over the max of %d - scaling down to meet the max", numPods, args.Max)
		}
	} else {
		logging.Logger.Tracef("Not scaling %s from %d pods", deployment.FriendlyName, numPods)
		scaleSizeGauge.Set(0)
		return nil
	}

	// Apply scale-down limits
	if podsToScaleTo < numPods {
		now := time.Now()
		nextAllowedScaleDown := lastScaleDown.Add(args.ScaleDown.Delay)
		if now.Before(nextAllowedScaleDown) {
			logging.Logger.Debugf("Not scaling down %s from %d to %d pods - cannot scale down until %s", deployment.FriendlyName, numPods, podsToScaleTo, nextAllowedScaleDown.String())
			scaleDownLimitedCounter.Inc()
			scaleSizeGauge.Set(0)
			return nil
		}

		podsToScaleToMin := numPods - args.ScaleDown.Max
		if podsToScaleTo < podsToScaleToMin {
			logging.Logger.Debugf("Capping the scale down from %d to %d pods", podsToScaleTo, podsToScaleToMin)
			podsToScaleTo = podsToScaleToMin
		}
	}

	if numPods != podsToScaleTo {
		// Apply metrics
		if podsToScaleTo < numPods {
			scaleDownCounter.Inc()
		} else {
			scaleUpCounter.Inc()
		}
		scaleSizeGauge.Set(float64(podsToScaleTo - numPods))

		logging.Logger.Infof("Scaling %s from %d to %d pods", deployment.FriendlyName, numPods, podsToScaleTo)
		err := k8sClient.Sync().Scale(deployment, podsToScaleTo)
		if err == nil && scale < 0 {
			lastScaleDown = time.Now()
		}
		return err
	}

	scaleSizeGauge.Set(0)

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

func getActiveAgentPodNames(agents []azuredevops.AgentDetails, podNames collections.StringSet) collections.StringSet {
	activeAgentPodNames := make(collections.StringSet)
	for _, agent := range agents {
		if strings.EqualFold(agent.Status, "online") {
			podName := agent.SystemCapabilities["HOSTNAME"]
			if podNames.Contains(podName) && agent.AssignedRequest != nil {
				activeAgentPodNames.Add(podName)
			}
		}
	}
	return activeAgentPodNames
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
