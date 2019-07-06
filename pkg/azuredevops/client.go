package azuredevops

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

const getPoolsEndpoint = "/_apis/distributedtask/pools"

// Parameter 1 is the Pool ID
const getPoolAgentsEndpoint = "/_apis/distributedtask/pools/%d/agents?includeCapabilities=true&includeAssignedRequest=true&includeLastCompletedRequest=true"

// Parameter 1 is the Pool ID, Parameter 2 is the Agent ID
const getAgentEndpoint = "/_apis/distributedtask/pools/%d/agents/%d?includeCapabilities=true&includeAssignedRequest=true&includeLastCompletedRequest=true"

const getPoolJobRequestsEndpoint = "/_apis/distributedtask/pools/%d/jobrequests"

const acceptHeader = "application/json;api-version=5.0-preview.1"

// Client is used to call Azure Devops
type Client struct {
	baseURL string

	token string
}

// MakeClient creates a new Azure Devops client
func MakeClient(baseURL string, token string) Client {
	if !strings.HasSuffix(baseURL, "") {
		baseURL = strings.TrimSuffix(baseURL, "/")
	}
	return Client{
		baseURL: baseURL,
		token:   token,
	}
}

func (c *Client) executeGETRequest(endpoint string, response interface{}) error {
	request, err := http.NewRequest("GET", c.baseURL+endpoint, nil)

	if err != nil {
		return err
	}

	request.Header.Set("Accept", acceptHeader)
	request.Header.Set("User-Agent", "go-azp-agent-autoscaler")

	request.SetBasicAuth("user", c.token)

	httpClient := http.Client{}
	httpResponse, err := httpClient.Do(request)
	if err != nil {
		return err
	}

	defer httpResponse.Body.Close()

	if httpResponse.StatusCode != 200 {
		return NewHTTPError(httpResponse)
	}

	err = json.NewDecoder(httpResponse.Body).Decode(response)
	if err != nil {
		return fmt.Errorf("Error - could not parse JSON response from %s: %s", endpoint, err.Error())
	}

	return nil
}

// PoolDetailsResponse is a wrapper for []PoolDetails to allow also returning an error in channels
type PoolDetailsResponse struct {
	Pools []PoolDetails
	Err   error
}

// ListPools retrieves a list of agent pools
func (c *Client) ListPools(channel chan<- PoolDetailsResponse) {
	response := new(PoolList)
	err := c.executeGETRequest(getPoolsEndpoint, response)
	if err != nil {
		channel <- PoolDetailsResponse{nil, err}
	} else {
		channel <- PoolDetailsResponse{response.Value, nil}
	}
}

// PoolAgentsResponse is a wrapper for []AgentDetails to allow also returning an error in channels
type PoolAgentsResponse struct {
	Agents []AgentDetails
	Err    error
}

// ListPoolAgents retrieves all of the agents in a pool
func (c *Client) ListPoolAgents(channel chan<- PoolAgentsResponse, poolID int) {
	response := new(Pool)
	endpoint := fmt.Sprintf(getPoolAgentsEndpoint, poolID)
	err := c.executeGETRequest(endpoint, response)
	if err != nil {
		channel <- PoolAgentsResponse{nil, err}
	} else {
		channel <- PoolAgentsResponse{response.Value, nil}
	}
}

// PoolAgentResponse is a wrapper for AgentDetails to allow also returning an error in channels
type PoolAgentResponse struct {
	Agent *AgentDetails
	Err   error
}

// GetPoolAgent retrieves a single agent in a pool
func (c *Client) GetPoolAgent(channel chan<- PoolAgentResponse, poolID int, agentID int) {
	response := new(AgentDetails)
	endpoint := fmt.Sprintf(getAgentEndpoint, poolID, agentID)
	err := c.executeGETRequest(endpoint, response)
	if err != nil {
		channel <- PoolAgentResponse{nil, err}
	} else {
		channel <- PoolAgentResponse{response, nil}
	}
}

// JobRequestsResponse is a wrapper for JobRequests to allow also returning an error in channels
type JobRequestsResponse struct {
	Jobs []JobRequest
	Err  error
}

// ListJobRequests retrieves the job requests for a pool
func (c *Client) ListJobRequests(channel chan<- JobRequestsResponse, poolID int) {
	response := new(JobRequests)
	endpoint := fmt.Sprintf(getPoolJobRequestsEndpoint, poolID)
	err := c.executeGETRequest(endpoint, response)
	if err != nil {
		channel <- JobRequestsResponse{nil, err}
	} else {
		channel <- JobRequestsResponse{response.Value, nil}
	}
}
