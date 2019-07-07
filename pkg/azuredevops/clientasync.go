package azuredevops

import (
	"strings"
)

// ClientAsync is an async version of Client
type ClientAsync interface {
	ListPoolsAsync(channel chan<- PoolDetailsResponse)
	ListPoolsByNameAsync(channel chan<- PoolDetailsResponse, poolName string)
	ListPoolAgentsAsync(channel chan<- PoolAgentsResponse, poolID int)
	GetPoolAgentAsync(channel chan<- PoolAgentResponse, poolID int, agentID int)
	ListJobRequestsAsync(channel chan<- JobRequestsResponse, poolID int)
}

// ClientAsyncImpl is the async interface implementation that calls Azure Devops
type ClientAsyncImpl struct {
	client Client
}

// MakeClient creates a new Azure Devops client
func MakeClient(baseURL string, token string) ClientAsync {
	if !strings.HasSuffix(baseURL, "") {
		baseURL = strings.TrimSuffix(baseURL, "/")
	}
	return ClientAsyncImpl{
		client: ClientImpl{
			baseURL: baseURL,
			token:   token,
		},
	}
}

// PoolDetailsResponse is a wrapper for []PoolDetails to allow also returning an error in channels
type PoolDetailsResponse struct {
	Pools []PoolDetails
	Err   error
}

// ListPoolsAsync retrieves a list of agent pools
func (c ClientAsyncImpl) ListPoolsAsync(channel chan<- PoolDetailsResponse) {
	response, err := c.client.ListPools()
	channel <- PoolDetailsResponse{response, err}
}

// ListPoolsByNameAsync retrieves a list of agent pools with the given name
func (c ClientAsyncImpl) ListPoolsByNameAsync(channel chan<- PoolDetailsResponse, poolName string) {
	response, err := c.client.ListPoolsByName(poolName)
	channel <- PoolDetailsResponse{response, err}
}

// PoolAgentsResponse is a wrapper for []AgentDetails to allow also returning an error in channels
type PoolAgentsResponse struct {
	Agents []AgentDetails
	Err    error
}

// ListPoolAgentsAsync retrieves all of the agents in a pool
func (c ClientAsyncImpl) ListPoolAgentsAsync(channel chan<- PoolAgentsResponse, poolID int) {
	response, err := c.client.ListPoolAgents(poolID)
	channel <- PoolAgentsResponse{response, err}
}

// PoolAgentResponse is a wrapper for AgentDetails to allow also returning an error in channels
type PoolAgentResponse struct {
	Agent *AgentDetails
	Err   error
}

// GetPoolAgentAsync retrieves a single agent in a pool
func (c ClientAsyncImpl) GetPoolAgentAsync(channel chan<- PoolAgentResponse, poolID int, agentID int) {
	response, err := c.client.GetPoolAgent(poolID, agentID)
	channel <- PoolAgentResponse{response, err}
}

// JobRequestsResponse is a wrapper for JobRequests to allow also returning an error in channels
type JobRequestsResponse struct {
	Jobs []JobRequest
	Err  error
}

// ListJobRequestsAsync retrieves the job requests for a pool
func (c ClientAsyncImpl) ListJobRequestsAsync(channel chan<- JobRequestsResponse, poolID int) {
	response, err := c.client.ListJobRequests(poolID)
	channel <- JobRequestsResponse{response, err}
}
