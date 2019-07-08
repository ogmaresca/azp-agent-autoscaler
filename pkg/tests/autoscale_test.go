package tests

import (
	"testing"

	"github.com/ggmaresca/azp-agent-autoscaler/pkg/args"
	"github.com/ggmaresca/azp-agent-autoscaler/pkg/azuredevops"
	"github.com/ggmaresca/azp-agent-autoscaler/pkg/kubernetes"
	"github.com/ggmaresca/azp-agent-autoscaler/pkg/scaling"
)

var (
	agentPoolID = 1
)

func TestAutoscale(t *testing.T) {
	var azdClient azuredevops.ClientAsync
	var k8sClient kubernetes.ClientAsync
	var deployment *kubernetes.Workload
	var args args.Args

	err := scaling.Autoscale(azdClient, agentPoolID, k8sClient, deployment, args)
	if err != nil {
		t.Error(err.Error())
	}
}
