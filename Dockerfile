FROM golang:1.12-alpine AS base

RUN apk update && apk add --no-cache git ca-certificates tzdata && update-ca-certificates

WORKDIR /go/src

ENV GO111MODULE on

RUN adduser -D -g '' azp-agent-autoscaler

COPY main.go /go/src/main.go

COPY go.mod /go/src/go.mod

#COPY pkg /go/src/github.com/ggmaresca/azp-agent-autoscaler/pkg

COPY pkg /go/src/pkg

RUN go get -d

#RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -a -installsuffix cgo -o /go/bin/azp-agent-autoscaler .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /go/bin/azp-agent-autoscaler

#RUN chown azp-agent-autoscaler /go/bin/azp-agent-autoscaler

#RUN chmod +x /go/bin/azp-agent-autoscaler

#FROM scratch

#COPY --from=base /usr/share/zoneinfo /usr/share/zoneinfo
#COPY --from=base /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
#COPY --from=base /etc/passwd /etc/passwd

#COPY --from=base /bin/sh /bin/sh

#COPY --from=base /go/bin /go/bin/azp-agent-autoscaler

#RUN chmod +x /bin/azp-agent-autoscaler

#USER azp-agent-autoscaler

ENTRYPOINT ["/go/bin/azp-agent-autoscaler"]
