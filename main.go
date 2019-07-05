package main

import (
	"flag"
	"fmt"
	"os"
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

	logger = log.Logger{
		Out: os.Stderr,
		Formatter: &log.TextFormatter{
			DisableColors: true,
			FullTimestamp: true,
		},
		Hooks:        make(log.LevelHooks),
		Level:        log.InfoLevel,
		ExitFunc:     os.Exit,
		ReportCaller: false,
	}
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
	if rate.Seconds() <= 1 {
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

	logger.SetLevel(logrusLevel)

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
	go azdClient.ListPools(agentPoolsChan)

	deployment := <-deploymentChan
	if deployment.Err != nil {
		logger.Panicf("Error retrieving %s/%s in namespace %s: %s", strings.ToLower(*resourceType), *resourceName, *resourceNamespace, deployment.Err.Error())
	}
	if err := <-verifyHPAChan; err != nil {
		logger.Panic(err.Error())
	}
	agentPools := <-agentPoolsChan
	if agentPools.Err != nil {
		logger.Panicf("Error retrieving agent pools: %s", agentPools.Err.Error())
	} else if len(agentPools.Pools) == 0 {
		logger.Panic("Error - did not find any agent pools")
	}

	// Discover the pool name from the environment variables
	var agentPoolName *string
agentPoolNameLoop:
	for _, container := range deployment.Resource.PodTemplateSpec.Spec.Containers {
		for _, env := range container.Env {
			if env.Name == poolNameEnvVar {
				envValue, err := helpers.GetEnvValue(env)
				if err != nil {
					logger.Panicf("Error getting Agent Pool - could not retrieve environment variable %s from statefulset/%s: %s", poolNameEnvVar, deployment.Resource.Name, err.Error())
				}
				agentPoolName = &envValue
				break agentPoolNameLoop
			}
		}
	}
	if agentPoolName == nil {
		logger.Panicf("Could not retrieve environment variable %s from statefulset/%s", poolNameEnvVar, deployment.Resource.Name)
	} else {
		logger.Debugf("Found agent pool %s from %s", *agentPoolName, deployment.Resource.FriendlyName)
	}

	var agentPoolID *int
	for _, agentPool := range agentPools.Pools {
		if !agentPool.IsHosted && agentPool.Name == *agentPoolName {
			agentPoolID = &agentPool.ID
			break
		}
	}
	if agentPoolID == nil {
		logger.Panicf("Error - could not find an agent pool with name %s", *agentPoolName)
	} else {
		logger.Debugf("Agent pool %s has ID %d", *agentPoolName, *agentPoolID)
	}

	getAgentsFunc := func(channel chan<- azuredevops.PoolAgentsResponse) {
		azdClient.ListPoolAgents(channel, *agentPoolID)
	}

	for {
		err := autoscale(getAgentsFunc, deployment.Resource)
		if err != nil {
			logger.Panicf("Error autoscaling statefulset/%s: %s", deployment.Resource.Name, err.Error())
		}

		time.Sleep(*rate)
	}

	logger.Info("Exiting azp-agent-autoscaler")
}

func autoscale(getAgentsFunc func(channel chan<- azuredevops.PoolAgentsResponse), deployment *helpers.KubernetesWorkload) error {
	agentsChan := make(chan azuredevops.PoolAgentsResponse)
	podsChan := make(chan helpers.Pods)

	go getAgentsFunc(agentsChan)
	go helpers.GetPods(podsChan, deployment)

	agents := <-agentsChan
	if agents.Err != nil {
		return agents.Err
	}
	pods := <-podsChan
	if pods.Err != nil {
		return pods.Err
	}

	podNames := make(map[string]struct{})
	numPods := int32(len(pods.Pods))
	numRunningPods, numPendingPods := int32(0), int32(0)
	for _, pod := range pods.Pods {
		var void struct{}
		podNames[pod.Name] = void
		if pod.Status.Phase == corev1.PodRunning {
			numRunningPods = numRunningPods + 1
		} else if pod.Status.Phase == corev1.PodPending {
			numPendingPods = numPendingPods + 1
		}
	}
	numFailedPods := numPods - numRunningPods - numPendingPods

	logger.Tracef("%d pods (%d running, %d pending, %d failed)", numPods, numRunningPods, numPendingPods, numFailedPods)

	min32 := int32(*min)
	max32 := int32(*max)

	// Get number of active agents
	numActiveAgents := int32(0)
	for _, agent := range agents.Agents {
		if strings.EqualFold(agent.Status, "online") {
			podName := agent.SystemCapabilities["HOSTNAME"]
			_, inCluster := podNames[podName]
			if inCluster && agent.AssignedRequest != nil {
				numActiveAgents = numActiveAgents + 1
			}
		}
	}

	logger.Debugf("Found %d active agents out of %d agents in the cluster.", numActiveAgents, numPods)

	if numRunningPods != numPods {
		logger.Infof("Not scaling - there are %d pending pods and %d failed pods.", numPendingPods, numFailedPods)
		return nil
	}

	replicasToSet := min32
	if numActiveAgents > min32 {
		replicasToSet = numActiveAgents + min32
		if replicasToSet > max32 {
			// Because there's no built-in Max() function that takes ints...
			if numActiveAgents > max32 {
				replicasToSet = numActiveAgents
			} else {
				replicasToSet = max32
			}
		}
	}

	return helpers.Scale(deployment, replicasToSet)
}
