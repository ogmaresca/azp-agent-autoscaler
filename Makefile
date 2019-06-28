go-build:
	go build ./...

docker-build:
	docker build -t azp-agent-autoscaler:dev .
