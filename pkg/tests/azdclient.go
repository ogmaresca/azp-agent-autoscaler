package tests

import (
	"fmt"

	"github.com/ggmaresca/azp-agent-autoscaler/pkg/azuredevops"
)

// PoolDetails creates mock PoolDetails objects
func PoolDetails(num int) []azuredevops.PoolDetails {
	var pools []azuredevops.PoolDetails
	for i := 0; i < num; i++ {
		pools = append(pools, azuredevops.PoolDetails{
			Definition: azuredevops.Definition{
				ID:   i,
				Name: fmt.Sprintf("pool-%d", i),
			},
			IsHosted: false,
		})
	}
	return pools
}

// Agents creates mock AgentDetails objects
func Agents(num int, free bool) []azuredevops.AgentDetails {
	var agents []azuredevops.AgentDetails
	for i := 0; i < num; i++ {
		agent := azuredevops.AgentDetails{
			Agent: azuredevops.Agent{
				Definition: azuredevops.Definition{
					ID:   i,
					Name: fmt.Sprintf("agent-%d", i),
				},
				Status: "online",
			},
			SystemCapabilities: map[string]string{
				"HOSTNAME": fmt.Sprintf("agent-%d", i),
			},
		}
		if !free {
			agent.AssignedRequest = &Jobs(1, false, []azuredevops.AgentDetails{agent})[0]
		}
		agents = append(agents, agent)
	}
	return agents
}

// Jobs creates mock JobRequest objects
func Jobs(num int, queued bool, agents []azuredevops.AgentDetails) []azuredevops.JobRequest {
	var jobs []azuredevops.JobRequest
	baseAgents := []azuredevops.Agent{}
	for _, agent := range agents {
		baseAgents = append(baseAgents, agent.Agent)
	}
	for i := 0; i < num; i++ {
		job := azuredevops.JobRequest{
			MatchesAllAgentsInPool: true,
			MatchedAgents:          baseAgents,
		}
		if !queued {
			job.ReservedAgent = &baseAgents[0]
		}
		jobs = append(jobs, job)
	}
	return jobs
}

// MockAZDClient is a mock Azure Devops client
type MockAZDClient struct {
	NumPools         int
	ErrorListPools   bool
	NumFreeAgents    int
	NumRunningAgents int
	ErrorAgents      bool
	NumQueuedJobs    int
	NumRunningJobs   int
	ErrorJobs        bool
}

// ListPoolsAsync retrieves a list of agent pools
func (c MockAZDClient) ListPoolsAsync(channel chan<- azuredevops.PoolDetailsResponse) {
	if c.ErrorListPools {
		channel <- azuredevops.PoolDetailsResponse{[]azuredevops.PoolDetails{}, fmt.Errorf("Mock AZD Client Error")}
	} else {
		channel <- azuredevops.PoolDetailsResponse{PoolDetails(c.NumPools), nil}
	}
}

// ListPoolsByNameAsync retrieves a list of agent pools with the given name
func (c MockAZDClient) ListPoolsByNameAsync(channel chan<- azuredevops.PoolDetailsResponse, poolName string) {
	if c.ErrorListPools {
		channel <- azuredevops.PoolDetailsResponse{[]azuredevops.PoolDetails{}, fmt.Errorf("Mock AZD Client Error")}
	} else {
		pools := PoolDetails(c.NumPools)
		for _, pool := range pools {
			if pool.Name == poolName {
				channel <- azuredevops.PoolDetailsResponse{[]azuredevops.PoolDetails{pool}, nil}
				return
			}
		}
		channel <- azuredevops.PoolDetailsResponse{[]azuredevops.PoolDetails{}, nil}
	}
}

// ListPoolAgentsAsync retrieves all of the agents in a pool
func (c MockAZDClient) ListPoolAgentsAsync(channel chan<- azuredevops.PoolAgentsResponse, poolID int) {
	if c.ErrorListPools {
		channel <- azuredevops.PoolAgentsResponse{[]azuredevops.AgentDetails{}, fmt.Errorf("Mock AZD Client Error")}
	} else {
		agents := Agents(c.NumRunningAgents, false)
		agents = append(agents, Agents(c.NumFreeAgents, false)...)
		channel <- azuredevops.PoolAgentsResponse{agents, nil}
	}
}

// ListJobRequestsAsync retrieves the job requests for a pool
func (c MockAZDClient) ListJobRequestsAsync(channel chan<- azuredevops.JobRequestsResponse, poolID int) {
	if c.ErrorListPools {
		channel <- azuredevops.JobRequestsResponse{[]azuredevops.JobRequest{}, fmt.Errorf("Mock AZD Client Error")}
	} else {
		agents := Agents(c.NumRunningAgents, false)
		agents = append(agents, Agents(c.NumFreeAgents, false)...)
		jobs := Jobs(c.NumRunningJobs, false, agents)
		jobs = append(jobs, Jobs(c.NumFreeAgents, true, agents)...)
		channel <- azuredevops.JobRequestsResponse{jobs, nil}
	}
}
