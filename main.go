package main

import (
	"flag"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/ggmaresca/azp-agent-autoscaler/pkg/azuredevops"
	"github.com/ggmaresca/azp-agent-autoscaler/pkg/helpers"
)

const poolNameEnvVar = "AZP_POOL"

func main() {
	// Parse arguments
	minStr := flag.String("min", "1", "Minimum number of free agents to keep alive. Minimum of 1.")
	maxStr := flag.String("max", "100", "Maximum number of agents allowed.")
	rateStr := flag.String("rate", "10s", "Duration to check the number of agents.")
	resourceType := flag.String("type", "StatefulSet", "Resource type of the agent. Only StatefulSet is supported.")
	resourceName := flag.String("name", "", "The name of the StatefulSet.")
	resourceNamespace := flag.String("namespace", "", "The namespace of the StatefulSet.")
	azpToken := flag.String("token", "", "The Azure Devops token.")
	azpURL := flag.String("url", "", "The Azure Devops URL. https://dev.azure.com/AccountName")

	flag.Parse()

	// Validate arguments
	min, err := strconv.ParseInt(*minStr, 10, 32)
	if err != nil {
		panic(fmt.Sprintf("Error converting min argument to int: %s", err.Error()))
	} else if min < 1 {
		panic("Error - min argument cannot be less than 1.")
	}

	max, err := strconv.ParseInt(*maxStr, 10, 32)
	if err != nil {
		panic(fmt.Sprintf("Error converting max argument to int: %s", err.Error()))
	} else if max <= min {
		panic("Error - max pods argument must be greater than the minimum.")
	}

	rate, err := time.ParseDuration(*rateStr)
	if err != nil {
		panic(fmt.Sprintf("Error parsing rate: %s", err.Error()))
	} else if rate.Seconds() <= 1 {
		panic(fmt.Sprintf("Error - rate '%s' is too low.", rate.String()))
	}

	if *resourceType != "StatefulSet" {
		panic(fmt.Sprintf("Error - Unknown resource type %s", *resourceType))
	}

	if *resourceName == "" {
		panic(fmt.Sprintf("Error - %s name is required.", *resourceType))
	}

	if *resourceNamespace == "" {
		panic("Error - namespace is required.")
	}

	if *azpToken == "" {
		panic("Error - the Azure Devops token is required.")
	}

	if *azpURL == "" {
		panic("Error - the Azure Devops URL is required.")
	}

	// Initialize Azure Devops client
	azdClient := azuredevops.MakeClient(*azpURL, *azpToken)

	deploymentChan := make(chan helpers.KubernetesResourceReturn)
	verifyHPAChan := make(chan error)
	agentPoolsChan := make(chan azuredevops.PoolDetailsResponse)

	// Get AZP agent deployment
	go helpers.GetK8sResource(deploymentChan, *resourceType, *resourceNamespace, *resourceName)
	// Verify there isn't a HorizontalPodAutoscaler
	go helpers.VerifyNoHorizontalPodAutoscaler(verifyHPAChan, *resourceType, *resourceNamespace, *resourceName)
	// Get all agent pools
	go azdClient.ListPools(agentPoolsChan)

	deployment := <-deploymentChan
	if deployment.Err != nil {
		panic(fmt.Sprintf("Error retrieving %s/%s in namespace %s: %s", strings.ToLower(*resourceType), *resourceName, *resourceNamespace, err.Error()))
	}
	if err = <-verifyHPAChan; err != nil {
		panic(err.Error())
	}
	agentPools := <-agentPoolsChan
	if agentPools.Err != nil {
		panic(fmt.Sprintf("Error retrieving agent pools: %s", err.Error()))
	} else if len(agentPools.Pools) == 0 {
		panic("Error - did not find any agent pools")
	}

	// Discover the pool name from the environment variables
	var agentPoolName *string
agentPoolNameLoop:
	for _, container := range deployment.Resource.PodTemplateSpec.Spec.Containers {
		for _, env := range container.Env {
			if env.Name == poolNameEnvVar {
				envValue, err := helpers.GetEnvValue(env)
				if err != nil {
					panic(fmt.Sprintf("Error getting Agent Pool - could not retrieve environment variable %s from statefulset/%s: %s", poolNameEnvVar, deployment.Resource.Name, err.Error()))
				}
				agentPoolName = &envValue
				break agentPoolNameLoop
			}
		}
	}
	if agentPoolName == nil {
		panic(fmt.Sprintf("Could not retrieve environment variable %s from statefulset/%s", poolNameEnvVar, deployment.Resource.Name))
	}

	var agentPoolID *int
	for _, agentPool := range agentPools.Pools {
		if agentPool.Name == *agentPoolName {
			agentPoolID = &agentPool.ID
		}
	}
	if agentPoolID == nil {
		panic(fmt.Sprintf("Error - could not find an agent pool with name %s", *agentPoolName))
	}

	getAgentsFunc := func(channel chan<- azuredevops.PoolAgentsResponse) {
		azdClient.ListPoolAgents(channel, *agentPoolID)
	}

	for {
		err = autoscale(getAgentsFunc, deployment.Resource, int32(min), int32(max))
		if err != nil {
			panic(fmt.Sprintf("Error autoscaling statefulset/%s: %s", deployment.Resource.Name, err.Error()))
		}

		time.Sleep(rate)
	}

	println("Exiting azp-agent-autoscaler")
}

func autoscale(getAgentsFunc func(channel chan<- azuredevops.PoolAgentsResponse), deployment *helpers.KubernetesResource, min int32, max int32) error {
	agentsChan := make(chan azuredevops.PoolAgentsResponse)
	go getAgentsFunc(agentsChan)
	agents := <-agentsChan
	if agents.Err != nil {
		return agents.Err
	}

	replicasToSet := min

	numActiveAgents := int32(0)
	for _, agent := range agents.Agents {
		if agent.AssignedRequest != nil {
			numActiveAgents = numActiveAgents + 1
		}
	}

	// TODO implement log framework
	// TODO take into account only active pods when determine to scale up/down
	// TODO can the agents be filtered to only include those in the statefulset? HOSTNAME is best bet
	println(fmt.Sprintf("Found %d active agents out of %d", numActiveAgents, len(agents.Agents)))

	// Get number of active agents
	if numActiveAgents > min {
		replicasToSet = numActiveAgents + min
		if replicasToSet > max {
			// Because there's no built-in Max method that takes ints...
			if numActiveAgents > max {
				replicasToSet = numActiveAgents
			} else {
				replicasToSet = max
			}
		}
	}

	return helpers.Scale(deployment, replicasToSet)
}
