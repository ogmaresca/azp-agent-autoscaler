package main

import (
	"flag"
	"fmt"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	corev1 "k8s.io/api/core/v1"

	"github.com/ggmaresca/azp-agent-autoscaler/pkg/azuredevops"
	"github.com/ggmaresca/azp-agent-autoscaler/pkg/helpers"
)

const poolNameEnvVar = "AZP_POOL"

var (
	logLevel          = flag.String("log-level", "info", "Log level (trace, debug, info, warn, error, fatal, panic).")
	min               = flag.Int("min", 1, "Minimum number of free agents to keep alive. Minimum of 1.")
	max               = flag.Int("max", 100, "Maximum number of agents allowed.")
	rate              = flag.Duration("rate", 10*time.Second, "Duration to check the number of agents.")
	scaleDownDelay    = flag.Duration("scale-down", 30*time.Second, "Wait time after scaling down to scale down again.")
	scaleDownMax      = flag.Int("scale-down-max", 1, "Maximum allowed number of pods to scale down.")
	resourceType      = flag.String("type", "StatefulSet", "Resource type of the agent. Only StatefulSet is supported.")
	resourceName      = flag.String("name", "", "The name of the StatefulSet.")
	resourceNamespace = flag.String("namespace", "", "The namespace of the StatefulSet.")
	azpToken          = flag.String("token", "", "The Azure Devops token.")
	azpURL            = flag.String("url", "", "The Azure Devops URL. https://dev.azure.com/AccountName")

	lastScaleDown = time.Date(1970, time.January, 1, 0, 0, 0, 0, time.UTC)
)

func main() {
	// Parse arguments
	flag.Parse()

	// Validate arguments
	var validationErrors []string
	logrusLevel, err := log.ParseLevel(*logLevel)
	if err != nil {
		validationErrors = append(validationErrors, err.Error())
	}
	if *min < 1 {
		validationErrors = append(validationErrors, "Min argument cannot be less than 1.")
	}
	if *max <= *min {
		validationErrors = append(validationErrors, "Max pods argument must be greater than the minimum.")
	}
	if rate == nil {
		validationErrors = append(validationErrors, "Rate is required.")
	} else if rate.Seconds() <= 1 {
		validationErrors = append(validationErrors, fmt.Sprintf("Rate '%s' is too low.", rate.String()))
	}
	if *scaleDownMax < 1 {
		validationErrors = append(validationErrors, fmt.Sprintf("Scale-down-max argument cannot be less than 1."))
	}
	if *resourceType != "StatefulSet" {
		validationErrors = append(validationErrors, fmt.Sprintf("Unknown resource type %s.", *resourceType))
	}
	if *resourceName == "" {
		validationErrors = append(validationErrors, fmt.Sprintf("%s name is required.", *resourceType))
	}
	if *resourceNamespace == "" {
		validationErrors = append(validationErrors, "Namespace is required.")
	}
	if *azpToken == "" {
		validationErrors = append(validationErrors, "The Azure Devops token is required.")
	}
	if *azpURL == "" {
		validationErrors = append(validationErrors, "The Azure Devops URL is required.")
	}
	if len(validationErrors) > 0 {
		panic(fmt.Errorf("Error(s) with arguments:\n%s", strings.Join(validationErrors, "\n")))
	}

	helpers.Logger.SetLevel(logrusLevel)

	// Initialize Azure Devops client
	azdClient := azuredevops.MakeClient(*azpURL, *azpToken)

	deploymentChan := make(chan helpers.KubernetesWorkloadReturn)
	verifyHPAChan := make(chan error)
	agentPoolsChan := make(chan azuredevops.PoolDetailsResponse)

	// Get AZP agent workload
	go helpers.GetK8sWorkload(deploymentChan, *resourceType, *resourceNamespace, *resourceName)
	// Verify there isn't a HorizontalPodAutoscaler
	go helpers.VerifyNoHorizontalPodAutoscaler(verifyHPAChan, *resourceType, *resourceNamespace, *resourceName)
	// Get all agent pools
	go azdClient.ListPoolsAsync(agentPoolsChan)

	deployment := <-deploymentChan
	if deployment.Err != nil {
		helpers.Logger.Panicf("Error retrieving %s/%s in namespace %s: %s", strings.ToLower(*resourceType), *resourceName, *resourceNamespace, deployment.Err.Error())
	}
	if err := <-verifyHPAChan; err != nil {
		helpers.Logger.Panic(err.Error())
	}
	agentPools := <-agentPoolsChan
	if agentPools.Err != nil {
		helpers.Logger.Panicf("Error retrieving agent pools: %s", agentPools.Err.Error())
	} else if len(agentPools.Pools) == 0 {
		helpers.Logger.Panic("Error - did not find any agent pools")
	}

	// Discover the pool name from the environment variables
	var agentPoolName *string
agentPoolNameLoop:
	for _, container := range deployment.Resource.PodTemplateSpec.Spec.Containers {
		for _, env := range container.Env {
			if env.Name == poolNameEnvVar {
				envValue, err := helpers.GetEnvValue(env)
				if err != nil {
					helpers.Logger.Panicf("Error getting Agent Pool - could not retrieve environment variable %s from statefulset/%s: %s", poolNameEnvVar, deployment.Resource.Name, err.Error())
				}
				agentPoolName = &envValue
				break agentPoolNameLoop
			}
		}
	}
	if agentPoolName == nil {
		helpers.Logger.Panicf("Could not retrieve environment variable %s from statefulset/%s", poolNameEnvVar, deployment.Resource.Name)
	} else {
		helpers.Logger.Debugf("Found agent pool %s from %s", *agentPoolName, deployment.Resource.FriendlyName)
	}

	var agentPoolID *int
	for _, agentPool := range agentPools.Pools {
		if !agentPool.IsHosted && agentPool.Name == *agentPoolName {
			agentPoolID = &agentPool.ID
			break
		}
	}
	if agentPoolID == nil {
		helpers.Logger.Panicf("Error - could not find an agent pool with name %s", *agentPoolName)
	} else {
		helpers.Logger.Debugf("Agent pool %s has ID %d", *agentPoolName, *agentPoolID)
	}

	getAgentsFunc := func(channel chan<- azuredevops.PoolAgentsResponse) {
		azdClient.ListPoolAgentsAsync(channel, *agentPoolID)
	}

	getJobRequestsFunc := func(channel chan<- azuredevops.JobRequestsResponse) {
		azdClient.ListJobRequestsAsync(channel, *agentPoolID)
	}

	for {
		err := autoscale(getAgentsFunc, getJobRequestsFunc, deployment.Resource)
		if err != nil {
			switch t := err.(type) {
			case azuredevops.HTTPError:
				httpError := err.(azuredevops.HTTPError)
				if httpError.RetryAfter != nil {
					helpers.Logger.Warnf("%s %s", t, httpError.Error())
					timeToSleep := httpError.RetryAfter
					if httpError.RetryAfter.Seconds() < rate.Seconds() {
						timeToSleep = rate
					}
					helpers.Logger.Infof("Retrying after %s", timeToSleep.String())
					time.Sleep(*timeToSleep)
				}
			default:
				// Do nothing
			}

			helpers.Logger.Panicf("Error autoscaling statefulset/%s: %s", deployment.Resource.Name, err.Error())
		} else {
			time.Sleep(*rate)
		}
	}

	helpers.Logger.Info("Exiting azp-agent-autoscaler")
}

func autoscale(getAgentsFunc func(channel chan<- azuredevops.PoolAgentsResponse), getJobRequestsFunc func(channel chan<- azuredevops.JobRequestsResponse), deployment *helpers.KubernetesWorkload) error {
	agentsChan := make(chan azuredevops.PoolAgentsResponse)
	jobsChan := make(chan azuredevops.JobRequestsResponse)
	podsChan := make(chan helpers.Pods)

	// Get all active agents
	go getAgentsFunc(agentsChan)
	// Get all queued jobs
	go getJobRequestsFunc(jobsChan)
	// Get all pods
	go helpers.GetPods(podsChan, deployment)

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
	podNames := make(helpers.StringSet)
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

	helpers.Logger.Tracef("%d pods (%d running, %d pending, %d failed)", numPods, numRunningPods, numPendingPods, numFailedPods)

	min32 := int32(*min)
	max32 := int32(*max)

	// Get number of active agents
	numActiveAgents := int32(0)
	activeAgentNames := make(helpers.StringSet)
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

	helpers.Logger.Debugf("Found %d active agents out of %d agents in the cluster. There are %d queued jobs.", numActiveAgents, numPods, numQueuedJobs)

	if numRunningPods != numPods {
		helpers.Logger.Infof("Not scaling - there are %d pending pods and %d failed pods.", numPendingPods, numFailedPods)
		return nil
	}

	// Determine delta for how much to scale by
	scale := int32(0)
	if numActiveAgents+min32 > numPods {
		// Scale up
		scale = numActiveAgents + min32 - numPods
		if numQueuedJobs > min32 {
			scale = scale + numQueuedJobs - min32
		}
	} else if numActiveAgents+min32+numQueuedJobs < numPods {
		scale = numPods - numActiveAgents - min32 - numQueuedJobs
	}

	// Apply scaling limits and scale down limits
	podsToScaleTo := numPods
	if scale > 0 {
		podsToScaleTo = helpers.MinInt32(max32, numPods+scale)
	} else if scale < 0 {
		podsToScaleTo = helpers.MaxInt32(min32, numPods+scale)
		scaleDownMax32 := int32(*scaleDownMax)

		now := time.Now()
		if now.Add(*scaleDownDelay).After(lastScaleDown) {
			helpers.Logger.Debugf("Not scaling down %s from %d to %d pods - cannot scale down until %s", deployment.FriendlyName, numPods, podsToScaleTo, now.Add(*scaleDownDelay).String())
		} else if numPods-podsToScaleTo > scaleDownMax32 {
			helpers.Logger.Debugf("Capping the scale down from %d to %d pods", podsToScaleTo, numPods-scaleDownMax32)
			podsToScaleTo = numPods - scaleDownMax32
		}
	} else {
		helpers.Logger.Tracef("Not scaling %s from %d pods", deployment.FriendlyName, numPods)
		return nil
	}

	helpers.Logger.Infof("Scaling %s from %d to %d pods", deployment.FriendlyName, numPods, podsToScaleTo)
	err := helpers.Scale(deployment, podsToScaleTo)
	if err == nil && scale < 0 {
		lastScaleDown = time.Now()
	}
	return err
}
