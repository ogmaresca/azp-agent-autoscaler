package azuredevops

// Agent is the the agent object returned in AgentRequest.ReservedAgent. It is also the base struct for AgentDetails.
type Agent struct {
	Definition
	Version           string `json:"version"`
	OSDescription     string `json:"osDescription"`
	Enabled           bool   `json:"enabled"`
	Status            string `json:"status"`
	ProvisioningState string `json:"provisioningState"`
	AccessPoint       string `json:"accessPoint"`
}

// AgentDetails is the response received when retrieving an individual agent.
// curl -u user:token https://dev.azure.com/organization/_apis/distributedtask/pools/9/agents/8?includeCapabilities=true&includeAssignedRequest=true&includeLastCompletedRequest=true'
type AgentDetails struct {
	Agent
	SystemCapabilities   map[string]string `json:"systemCapabilities"`
	MaxParallelism       int               `json:"maxParallelism"`
	CreatedOn            string            `json:"createdOn"`
	AssignedRequest      *JobRequest       `json:"assignedRequest"`
	LastCompletedRequest *JobRequest       `json:"lastCompletedRequest"`
}
