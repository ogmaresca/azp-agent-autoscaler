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
	AssignedRequest      AgentRequest      `json:"assignedRequest"`
	LastCompletedRequest AgentRequest      `json:"lastCompletedRequest"`
}

// AgentRequest represents a requested job a pool has or is running.
type AgentRequest struct {
	RequestID              int               `json:"requestId"`
	QueueTime              string            `json:"queueTime"`
	AssignTime             string            `json:"assignTime"`
	ReceiveTime            string            `json:"receiveTime"`
	FinishTime             string            `json:"finishTime"`
	Result                 string            `json:"result"`
	ServiceOwner           string            `json:"serviceOwner"`
	HostID                 string            `json:"hostId"`
	ScopeID                string            `json:"scopeId"`
	PlanType               string            `json:"planType"`
	PlanID                 string            `json:"planId"`
	JobID                  string            `json:"jobId"`
	Demands                []string          `json:"demands"`
	ReservedAgent          Agent             `json:"reservedAgent"`
	Data                   map[string]string `json:"data"`
	PoolID                 int               `json:"poolId"`
	OrchestrationID        string            `json:"orchestrationId"`
	MatchesAllAgentsInPool bool              `json:"matchesAllAgentsInPool"`
	Definition             Definition        `json:"definition"`
	Owner                  Definition        `json:"definition"`
}
