# Create base stage first to cache commands
FROM alpine:3.10 AS final_base

EXPOSE 10101

RUN apk update && apk add --no-cache ca-certificates bash

RUN adduser -D -g '' azp-agent-autoscaler

# Download modules
FROM golang:1.12-alpine3.10 AS base

RUN apk update && apk add --no-cache git ca-certificates tzdata

WORKDIR /go/src

ENV GO111MODULE on
ENV CGO_ENABLED 0
ENV GOOS linux
ENV GOARCH amd64

COPY go.mod /go/src/go.mod

RUN go mod download

# Compile
FROM base AS build

COPY main.go /go/src/main.go
COPY pkg /go/src/pkg

RUN go build -ldflags="-w -s" -o /go/bin/azp-agent-autoscaler

# Use alpine as final base stage
FROM final_base AS final

COPY --from=build /go/bin/azp-agent-autoscaler /bin/azp-agent-autoscaler

RUN chown azp-agent-autoscaler /bin/azp-agent-autoscaler

USER azp-agent-autoscaler

WORKDIR /home/azp-agent-autoscaler

ENTRYPOINT ["/bin/azp-agent-autoscaler"]
CMD []
