package main

import (
	"flag"
	"time"

	"github.com/ggmaresca/azp-agent-autoscaler/pkg/azuredevops"
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
		err := scaling.Autoscale(getAgentsFunc, getJobRequestsFunc, deployment.Resource, args)
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
