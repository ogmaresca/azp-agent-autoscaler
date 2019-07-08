#!/bin/bash

AZP_AGENT_AUTOSCALER_VERSION=$(cat version) && \
echo "Uploading azp-agent-autoscaler $AZP_AGENT_AUTOSCALER_VERSION" && \
docker tag azp-agent-autoscaler:dev docker.io/gmaresca/azp-agent-autoscaler:$AZP_AGENT_AUTOSCALER_VERSION && \
docker tag azp-agent-autoscaler:dev docker.io/gmaresca/azp-agent-autoscaler:latest && \
docker push docker.io/gmaresca/azp-agent-autoscaler:$AZP_AGENT_AUTOSCALER_VERSION && \
docker push docker.io/gmaresca/azp-agent-autoscaler:latest && \
docker rmi docker.io/gmaresca/azp-agent-autoscaler:$AZP_AGENT_AUTOSCALER_VERSION && \
docker rmi docker.io/gmaresca/azp-agent-autoscaler:latest && \
echo "Finished uploading azp-agent-autoscaler $AZP_AGENT_AUTOSCALER_VERSION"
