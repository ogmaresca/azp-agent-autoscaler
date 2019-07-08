go-build:
	go build -o ../bin/azp-agent-autoscaler .

go-run:
	../bin/azp-agent-autoscaler --name azp-agent --namespace default --token=${AZURE_DEVOPS_TOKEN} --url=${AZURE_DEVOPS_URL} --log-level=Trace

docker-build:
	docker build -t azp-agent-autoscaler:dev .

docker-run:
	docker run -it --rm --name=azp-agent-autoscaler -v ${HOME}/.kube:/home/azp-agent-autoscaler/.kube:ro --network=host azp-agent-autoscaler:dev --name=azp-agent --namespace=default --token=${AZURE_DEVOPS_TOKEN} --url=${AZURE_DEVOPS_URL} --log-level=Trace
