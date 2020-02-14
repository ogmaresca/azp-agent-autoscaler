# Download modules
FROM golang:1.13-alpine3.11 AS base

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
FROM alpine:3.11 AS final

EXPOSE 10101

RUN adduser -D -g '' -u 1000 azp-agent-autoscaler

USER 1000

WORKDIR /home/azp-agent-autoscaler

COPY --from=build --chown=azp-agent-autoscaler /go/bin/azp-agent-autoscaler /bin/azp-agent-autoscaler

ENTRYPOINT ["/bin/azp-agent-autoscaler"]
CMD []
