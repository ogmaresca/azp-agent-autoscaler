package azuredevops

// PoolList is the response received when listing pools.
// curl -u user:token https://dev.azure.com/organization/_apis/distributedtask/pools?api-version=5.0-preview.1
type PoolList struct {
	Count int           `json:"count"`
	Value []PoolDetails `json:"value"`
}

// PoolDetails is returned when listing agent pools.
type PoolDetails struct {
	Definition
	CreatedOn     string `json:"createdOn"`
	AutoProvision bool   `json:"autoProvision"`
	AutoSize      bool   `json:"autoSize"`
	TargetSize    int    `json:"targetSize"`
	AgentCloudID  int    `json:"agentCloudId"`
	CreatedBy     *User  `json:"createdBy"`
	Owner         *User  `json:"owner"`
	Scope         string `json:"scope"`
	IsHosted      bool   `json:"isHosted"`
	PoolType      string `json:"poolType"`
	Size          int    `json:"size"`
	IsLegacy      bool   `json:"isLegacy"`
}
