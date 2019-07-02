package azuredevops

// Pool is the response received when retrieving an individual pool.
// curl -u user:token https://dev.azure.com/organization/_apis/distributedtask/pools/9/agents?includeCapabilities=true&includeAssignedRequest=true&includeLastCompletedRequest=true'
type Pool struct {
	Count int            `json:"count"`
	Value []AgentDetails `json:"value"`
}
