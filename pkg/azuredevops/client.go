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
		return fmt.Errorf("Error - received HTTP status code %d when calling call to %s", httpResponse.StatusCode, endpoint)
	}

	err = json.NewDecoder(httpResponse.Body).Decode(response)
	if err != nil {
		return fmt.Errorf("Error - could not parse JSON response from %s: %s", endpoint, err.Error())
	}

	return nil
}

// ListPools retrieves a list of agent pools
func (c *Client) ListPools() ([]PoolDetails, error) {
	response := new(PoolList)
	err := c.executeGETRequest(getPoolsEndpoint, response)
	if err != nil {
		return nil, err
	}

	return response.Value, nil
}

// ListPoolAgents retrieves all of the agents in a pool
func (c *Client) ListPoolAgents(poolID int) ([]AgentDetails, error) {
	response := new(Pool)
	endpoint := fmt.Sprintf(getPoolAgentsEndpoint, poolID)
	err := c.executeGETRequest(endpoint, response)
	if err != nil {
		return nil, err
	}

	return response.Value, nil
}

// GetPoolAgent retrieves a single agent in a pool
func (c *Client) GetPoolAgent(poolID int, agentID int) (*AgentDetails, error) {
	response := new(AgentDetails)
	endpoint := fmt.Sprintf(getAgentEndpoint, poolID, agentID)
	err := c.executeGETRequest(endpoint, response)
	if err != nil {
		return nil, err
	}

	return response, nil
}
