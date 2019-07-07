package azuredevops

import "strings"

// JobRequests is the response received when retrieving a pool's jobs.
// curl -u user:token https://dev.azure.com/organization/_apis/distributedtask/pools/9/jobrequests'
type JobRequests struct {
	Count int          `json:"count"`
	Value []JobRequest `json:"value"`
}

// JobResult is the Result field of JobRequest
// Running or queued jobs don't have a result
type JobResult string

const (
	// JobResultFailed is the result for failed jobs
	JobResultFailed JobResult = "failed"
	// JobResultCanceled is the result for canceled jobs
	JobResultCanceled JobResult = "canceled"
	// JobResultSucceeded is the result for succeeded jobs
	JobResultSucceeded JobResult = "succeeded"
)

// JobRequest represents a requested job
type JobRequest struct {
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
	MatchedAgents          []Agent           `json:"matchedAgents"`
	ReservedAgent          *Agent            `json:"reservedAgent"`
	Data                   map[string]string `json:"data"`
	PoolID                 int               `json:"poolId"`
	OrchestrationID        string            `json:"orchestrationId"`
	MatchesAllAgentsInPool bool              `json:"matchesAllAgentsInPool"`
	Definition             *Definition       `json:"definition"`
	Owner                  *Definition       `json:"definition"`
	// AgentDelays            []struct{}        `json:"agentDelays"`
}

// IsQueuedOrRunning determines if a job is currently queued or running
func (j *JobRequest) IsQueuedOrRunning() bool {
	return !strings.EqualFold(j.Result, string(JobResultFailed)) &&
		!strings.EqualFold(j.Result, string(JobResultCanceled)) &&
		!strings.EqualFold(j.Result, string(JobResultSucceeded)) &&
		len(j.MatchedAgents) > 0
}
