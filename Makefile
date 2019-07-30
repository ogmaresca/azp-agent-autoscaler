go-build:
	go build -o ../bin/azp-agent-autoscaler .

go-run:
	../bin/azp-agent-autoscaler --name azp-agent --namespace default --token=${AZURE_DEVOPS_TOKEN} --url=${AZURE_DEVOPS_URL} --log-level=Trace

go-test:
	go test github.com/ggmaresca/azp-agent-autoscaler/pkg/tests

docker-build:
	docker build -t azp-agent-autoscaler:dev .

docker-run:
	docker run -it --rm --name=azp-agent-autoscaler -v ${HOME}/.kube:/home/azp-agent-autoscaler/.kube:ro --network=host azp-agent-autoscaler:dev --name=azp-agent --namespace=default --token=${AZURE_DEVOPS_TOKEN} --url=${AZURE_DEVOPS_URL} --log-level=Trace

docker-push:
	sh docker-push.sh

docker-clean:
	docker rmi azp-agent-autoscaler:dev

helm-lint:
	helm lint charts/azp-agent-autoscaler

helm-template:
	helm template charts/azp-agent-autoscaler --set azp.url=https://dev.azure.com/test,azp.token=abc123def456ghi789jkl,agents.name=azp-agent,pdb.enabled=true,serviceMonitor.enabled=true

helm-install:
	helm upgrade --debug --install azp-agent-autoscaler charts/azp-agent-autoscaler --set azp.url=${AZURE_DEVOPS_URL},azp.existingSecret=azp-agent,azp.existingSecretKey=azp-token,agents.name=azp-agent,logLevel=trace

helm-package:
	helm package charts/azp-agent-autoscaler -d charts && \
	helm repo index --merge charts/index.yaml charts
