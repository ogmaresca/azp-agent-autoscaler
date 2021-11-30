package tests

import (
	"fmt"

	"github.com/ogmaresca/azp-agent-autoscaler/pkg/azuredevops"
)

// PoolDetails creates mock PoolDetails objects
func PoolDetails(num int32, startPos int32) []azuredevops.PoolDetails {
	var pools []azuredevops.PoolDetails
	for i := startPos; i < startPos+num; i++ {
		pools = append(pools, azuredevops.PoolDetails{
			Definition: azuredevops.Definition{
				ID:   int(i),
				Name: fmt.Sprintf("pool-%d", i),
			},
			IsHosted: false,
		})
	}
	return pools
}

// Agents creates mock AgentDetails objects
func Agents(num int32, free bool, startPos int32) []azuredevops.AgentDetails {
	var agents []azuredevops.AgentDetails
	for i := startPos; i < startPos+num; i++ {
		agent := azuredevops.AgentDetails{
			Agent: azuredevops.Agent{
				Definition: azuredevops.Definition{
					ID:   int(i),
					Name: fmt.Sprintf("agent-%d", i),
				},
				Status: "online",
			},
			SystemCapabilities: map[string]string{
				"HOSTNAME": fmt.Sprintf("azp-agent-%d", i),
			},
		}
		if free {
			agent.AssignedRequest = nil
		} else {
			agent.AssignedRequest = &Jobs(1, false, []azuredevops.AgentDetails{agent}, 0, 0)[0]
		}
		agents = append(agents, agent)
	}
	return agents
}

// Jobs creates mock JobRequest objects
func Jobs(num int32, queued bool, agents []azuredevops.AgentDetails, startPos int32, runningAgentPos int32) []azuredevops.JobRequest {
	var jobs []azuredevops.JobRequest
	baseAgents := []azuredevops.Agent{}
	for _, agent := range agents {
		baseAgents = append(baseAgents, agent.Agent)
	}
	for i := startPos; i < startPos+num; i++ {
		job := azuredevops.JobRequest{
			JobID:                  fmt.Sprintf("job-%d-queued=%t", num, queued),
			MatchesAllAgentsInPool: true,
			MatchedAgents:          baseAgents,
		}
		if queued {
			job.ReservedAgent = nil
		} else {
			job.ReservedAgent = &baseAgents[runningAgentPos]
		}
		jobs = append(jobs, job)
	}
	return jobs
}

// mockAZDClient is a mock Azure Devops client
type mockAZDClient struct {
	NumPools         int32
	ErrorListPools   bool
	NumFreeAgents    int32
	NumRunningAgents int32
	ErrorAgents      bool
	NumQueuedJobs    int32
	ErrorJobs        bool
	FreeAgentsFirst  bool
}

// ListPoolsAsync retrieves a list of agent pools
func (c mockAZDClient) ListPoolsAsync(channel chan<- azuredevops.PoolDetailsResponse) {
	if c.ErrorListPools {
		channel <- azuredevops.PoolDetailsResponse{[]azuredevops.PoolDetails{}, fmt.Errorf("Mock AZD Client Error")}
	} else {
		channel <- azuredevops.PoolDetailsResponse{PoolDetails(c.NumPools, 0), nil}
	}
}

// ListPoolsByNameAsync retrieves a list of agent pools with the given name
func (c mockAZDClient) ListPoolsByNameAsync(channel chan<- azuredevops.PoolDetailsResponse, poolName string) {
	if c.ErrorListPools {
		channel <- azuredevops.PoolDetailsResponse{[]azuredevops.PoolDetails{}, fmt.Errorf("Mock AZD Client Error")}
	} else {
		pools := PoolDetails(c.NumPools, 0)
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
func (c mockAZDClient) ListPoolAgentsAsync(channel chan<- azuredevops.PoolAgentsResponse, poolID int) {
	if c.ErrorListPools {
		channel <- azuredevops.PoolAgentsResponse{[]azuredevops.AgentDetails{}, fmt.Errorf("Mock AZD Client Error")}
	} else {
		agents := c.listPoolAgents()
		channel <- azuredevops.PoolAgentsResponse{agents, nil}
	}
}

// ListJobRequestsAsync retrieves the job requests for a pool
func (c mockAZDClient) ListJobRequestsAsync(channel chan<- azuredevops.JobRequestsResponse, poolID int) {
	if c.ErrorListPools {
		channel <- azuredevops.JobRequestsResponse{[]azuredevops.JobRequest{}, fmt.Errorf("Mock AZD Client Error")}
	} else {
		agents := c.listPoolAgents()
		runningAgentPos := int32(0)
		if c.FreeAgentsFirst && c.NumRunningAgents > 0 {
			runningAgentPos = c.NumFreeAgents
		}
		jobs := Jobs(c.NumRunningAgents, false, agents, 0, runningAgentPos)
		jobs = append(jobs, Jobs(c.NumQueuedJobs, true, agents, int32(len(agents)), runningAgentPos)...)
		channel <- azuredevops.JobRequestsResponse{jobs, nil}
	}
}

func (c mockAZDClient) listPoolAgents() []azuredevops.AgentDetails {
	agents := []azuredevops.AgentDetails{}
	if c.FreeAgentsFirst {
		agents = append(agents, Agents(c.NumFreeAgents, true, 0)...)
		agents = append(agents, Agents(c.NumRunningAgents, false, int32(len(agents)))...)
	} else {
		agents = append(agents, Agents(c.NumRunningAgents, false, 0)...)
		agents = append(agents, Agents(c.NumFreeAgents, true, int32(len(agents)))...)
	}

	return agents
}
