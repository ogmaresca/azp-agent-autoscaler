package main

import (
	"flag"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"

	"github.com/ggmaresca/azp-agent-autoscaler/pkg/azuredevops"
	"github.com/ggmaresca/azp-agent-autoscaler/pkg/collections"
	"github.com/ggmaresca/azp-agent-autoscaler/pkg/kubernetes"
	"github.com/ggmaresca/azp-agent-autoscaler/pkg/logging"
	"github.com/ggmaresca/azp-agent-autoscaler/pkg/math"
	"github.com/ggmaresca/azp-agent-autoscaler/pkg/scaling"
)

const poolNameEnvVar = "AZP_POOL"

var (
	lastScaleDown = time.Date(1970, time.January, 1, 0, 0, 0, 0, time.UTC)
)

func main() {
	// Parse arguments
	flag.Parse()

	if err := scaling.ValidateArgs(); err != nil {
		panic(err.Error())
	}
	args := scaling.ArgsFromFlags()

	logging.Logger.SetLevel(args.Logging.Level)

	// Initialize Azure Devops client
	azdClient := azuredevops.MakeClient(args.AZD.URL, args.AZD.Token)

	deploymentChan := make(chan kubernetes.WorkloadReturn)
	verifyHPAChan := make(chan error)
	agentPoolsChan := make(chan azuredevops.PoolDetailsResponse)

	// Get AZP agent workload
	go kubernetes.GetK8sWorkload(deploymentChan, args.Kubernetes.Type, args.Kubernetes.Namespace, args.Kubernetes.Name)
	// Verify there isn't a HorizontalPodAutoscaler
	go kubernetes.VerifyNoHorizontalPodAutoscaler(verifyHPAChan, args.Kubernetes.Type, args.Kubernetes.Namespace, args.Kubernetes.Name)
	// Get all agent pools
	go azdClient.ListPoolsAsync(agentPoolsChan)

	deployment := <-deploymentChan
	if deployment.Err != nil {
		logging.Logger.Panicf("Error retrieving %s in namespace %s: %s", args.Kubernetes.FriendlyName(), args.Kubernetes.Namespace, deployment.Err.Error())
	}
	if err := <-verifyHPAChan; err != nil {
		logging.Logger.Panic(err.Error())
	}
	agentPools := <-agentPoolsChan
	if agentPools.Err != nil {
		logging.Logger.Panicf("Error retrieving agent pools: %s", agentPools.Err.Error())
	} else if len(agentPools.Pools) == 0 {
		logging.Logger.Panic("Error - did not find any agent pools")
	}

	// Discover the pool name from the environment variables
	var agentPoolName *string
agentPoolNameLoop:
	for _, container := range deployment.Resource.PodTemplateSpec.Spec.Containers {
		for _, env := range container.Env {
			if env.Name == poolNameEnvVar {
				envValue, err := kubernetes.GetEnvValue(env)
				if err != nil {
					logging.Logger.Panicf("Error getting Agent Pool - could not retrieve environment variable %s from statefulset/%s: %s", poolNameEnvVar, deployment.Resource.Name, err.Error())
				}
				agentPoolName = &envValue
				break agentPoolNameLoop
			}
		}
	}
	if agentPoolName == nil {
		logging.Logger.Panicf("Could not retrieve environment variable %s from statefulset/%s", poolNameEnvVar, deployment.Resource.Name)
	} else {
		logging.Logger.Debugf("Found agent pool %s from %s", *agentPoolName, deployment.Resource.FriendlyName)
	}

	var agentPoolID *int
	for _, agentPool := range agentPools.Pools {
		if !agentPool.IsHosted && agentPool.Name == *agentPoolName {
			agentPoolID = &agentPool.ID
			break
		}
	}
	if agentPoolID == nil {
		logging.Logger.Panicf("Error - could not find an agent pool with name %s", *agentPoolName)
	} else {
		logging.Logger.Debugf("Agent pool %s has ID %d", *agentPoolName, *agentPoolID)
	}

	getAgentsFunc := func(channel chan<- azuredevops.PoolAgentsResponse) {
		azdClient.ListPoolAgentsAsync(channel, *agentPoolID)
	}

	getJobRequestsFunc := func(channel chan<- azuredevops.JobRequestsResponse) {
		azdClient.ListJobRequestsAsync(channel, *agentPoolID)
	}

	for {
		err := autoscale(getAgentsFunc, getJobRequestsFunc, deployment.Resource, args)
		if err != nil {
			switch t := err.(type) {
			case azuredevops.HTTPError:
				httpError := err.(azuredevops.HTTPError)
				if httpError.RetryAfter != nil {
					logging.Logger.Warnf("%s %s", t, httpError.Error())
					timeToSleep := math.MaxDuration(*httpError.RetryAfter, args.Rate)
					logging.Logger.Infof("Retrying after %s", timeToSleep.String())
					time.Sleep(timeToSleep)
				}
			default:
				// Do nothing
			}

			logging.Logger.Panicf("Error autoscaling statefulset/%s: %s", deployment.Resource.Name, err.Error())
		} else {
			time.Sleep(args.Rate)
		}
	}

	logging.Logger.Info("Exiting azp-agent-autoscaler")
}

func autoscale(getAgentsFunc func(channel chan<- azuredevops.PoolAgentsResponse), getJobRequestsFunc func(channel chan<- azuredevops.JobRequestsResponse), deployment *kubernetes.Workload, args scaling.Args) error {
	agentsChan := make(chan azuredevops.PoolAgentsResponse)
	jobsChan := make(chan azuredevops.JobRequestsResponse)
	podsChan := make(chan kubernetes.Pods)

	// Get all active agents
	go getAgentsFunc(agentsChan)
	// Get all queued jobs
	go getJobRequestsFunc(jobsChan)
	// Get all pods
	go kubernetes.GetPods(podsChan, deployment)

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
	numRunningPods, numPendingPods := int32(0), int32(0)
	for _, pod := range pods.Pods {
		podNames.Add(pod.Name)
		if pod.Status.Phase == corev1.PodRunning {
			numRunningPods = numRunningPods + 1
		} else if pod.Status.Phase == corev1.PodPending {
			numPendingPods = numPendingPods + 1
		}
	}
	numFailedPods := numPods - numRunningPods - numPendingPods

	logging.Logger.Tracef("%d pods (%d running, %d pending, %d failed)", numPods, numRunningPods, numPendingPods, numFailedPods)

	// Get number of active agents
	numActiveAgents := int32(0)
	activeAgentNames := make(collections.StringSet)
	for _, agent := range agents.Agents {
		if strings.EqualFold(agent.Status, "online") {
			podName := agent.SystemCapabilities["HOSTNAME"]
			if podNames.Contains(podName) && agent.AssignedRequest != nil {
				numActiveAgents = numActiveAgents + 1
				activeAgentNames.Add(agent.Name)
			}
		}
	}

	// Determine the number of jobs that are queued
	numQueuedJobs := int32(0)
	for _, job := range jobs.Jobs {
		if job.IsQueuedOrRunning() {
			if job.ReservedAgent == nil && len(job.MatchedAgents) > 0 {
				for _, agent := range job.MatchedAgents {
					if activeAgentNames.Contains(agent.Name) {
						numQueuedJobs = numQueuedJobs + 1
						break
					}
				}
			}
		}
	}

	logging.Logger.Debugf("Found %d active agents out of %d agents in the cluster. There are %d queued jobs.", numActiveAgents, numPods, numQueuedJobs)

	if numRunningPods != numPods {
		logging.Logger.Infof("Not scaling - there are %d pending pods and %d failed pods.", numPendingPods, numFailedPods)
		return nil
	}

	// Determine delta for how much to scale by
	scale := int32(0)
	if numActiveAgents+args.Min > numPods {
		// Scale up
		scale = numActiveAgents + args.Min - numPods
		if numQueuedJobs > args.Min {
			scale = scale + numQueuedJobs - args.Min
		}
	} else if numActiveAgents+args.Min+numQueuedJobs < numPods {
		scale = numPods - numActiveAgents - args.Min - numQueuedJobs
	}

	// Apply scaling limits and scale down limits
	podsToScaleTo := numPods
	if scale > 0 {
		podsToScaleTo = math.MinInt32(args.Max, numPods+scale)
	} else if scale < 0 {
		podsToScaleTo = math.MaxInt32(args.Min, numPods+scale)

		now := time.Now()
		if now.Add(args.ScaleDown.Delay).After(lastScaleDown) {
			logging.Logger.Debugf("Not scaling down %s from %d to %d pods - cannot scale down until %s", deployment.FriendlyName, numPods, podsToScaleTo, now.Add(args.ScaleDown.Delay).String())
		} else if numPods-podsToScaleTo > args.ScaleDown.Max {
			logging.Logger.Debugf("Capping the scale down from %d to %d pods", podsToScaleTo, numPods-args.ScaleDown.Max)
			podsToScaleTo = numPods - args.ScaleDown.Max
		}
	} else {
		logging.Logger.Tracef("Not scaling %s from %d pods", deployment.FriendlyName, numPods)
		return nil
	}

	logging.Logger.Infof("Scaling %s from %d to %d pods", deployment.FriendlyName, numPods, podsToScaleTo)
	err := kubernetes.Scale(deployment, podsToScaleTo)
	if err == nil && scale < 0 {
		lastScaleDown = time.Now()
	}
	return err
}
