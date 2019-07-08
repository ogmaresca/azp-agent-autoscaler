package azuredevops

import (
	"encoding/json"
	"fmt"
	"net/http"
)

const getPoolsEndpoint = "/_apis/distributedtask/pools?poolName=%s"

// Parameter 1 is the Pool ID
const getPoolAgentsEndpoint = "/_apis/distributedtask/pools/%d/agents?includeCapabilities=true&includeAssignedRequest=true&includeLastCompletedRequest=true"

const getPoolJobRequestsEndpoint = "/_apis/distributedtask/pools/%d/jobrequests"

const acceptHeader = "application/json;api-version=5.0-preview.1"

// Client is used to call Azure Devops
type Client interface {
	ListPools() ([]PoolDetails, error)
	ListPoolsByName(poolName string) ([]PoolDetails, error)
	ListPoolAgents(poolID int) ([]AgentDetails, error)
	ListJobRequests(poolID int) ([]JobRequest, error)
}

// ClientImpl is the interface implementation that calls Azure Devops
type ClientImpl struct {
	baseURL string

	token string
}

func (c ClientImpl) executeGETRequest(endpoint string, response interface{}) error {
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

// ListPools retrieves a list of agent pools
func (c ClientImpl) ListPools() ([]PoolDetails, error) {
	return c.ListPoolsByName("")
}

// ListPoolsByName retrieves a list of agent pools with the given name
func (c ClientImpl) ListPoolsByName(poolName string) ([]PoolDetails, error) {
	response := new(PoolList)
	endpoint := fmt.Sprintf(getPoolsEndpoint, poolName)
	err := c.executeGETRequest(endpoint, response)
	if err != nil {
		return nil, err
	} else {
		return response.Value, nil
	}
}

// ListPoolAgents retrieves all of the agents in a pool
func (c ClientImpl) ListPoolAgents(poolID int) ([]AgentDetails, error) {
	response := new(Pool)
	endpoint := fmt.Sprintf(getPoolAgentsEndpoint, poolID)
	err := c.executeGETRequest(endpoint, response)
	if err != nil {
		return nil, err
	} else {
		return response.Value, nil
	}
}

// ListJobRequests retrieves the job requests for a pool
func (c ClientImpl) ListJobRequests(poolID int) ([]JobRequest, error) {
	response := new(JobRequests)
	endpoint := fmt.Sprintf(getPoolJobRequestsEndpoint, poolID)
	err := c.executeGETRequest(endpoint, response)
	if err != nil {
		return nil, err
	} else {
		return response.Value, nil
	}
}
