FROM golang:1.12-alpine3.10 AS base

RUN apk update && apk add --no-cache git ca-certificates tzdata

WORKDIR /go/src

ENV GO111MODULE on

COPY main.go /go/src/main.go
COPY go.mod /go/src/go.mod
COPY pkg /go/src/pkg

RUN go get -d

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /go/bin/azp-agent-autoscaler

FROM alpine:3.10

RUN apk update && apk add --no-cache ca-certificates bash

COPY --from=base /go/bin/azp-agent-autoscaler /bin/azp-agent-autoscaler

RUN adduser -D -g '' azp-agent-autoscaler
RUN chown azp-agent-autoscaler /bin/azp-agent-autoscaler

USER azp-agent-autoscaler

WORKDIR /home/azp-agent-autoscaler

ENTRYPOINT ["/bin/azp-agent-autoscaler"]
