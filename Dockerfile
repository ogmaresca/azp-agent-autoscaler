# Download trusted root certificates
FROM alpine:3.11 AS ca-certificates

RUN apk add --no-cache ca-certificates && \
    update-ca-certificates

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

# Use a distroless final base
FROM scratch AS final

EXPOSE 10101

USER 1000

WORKDIR /home/azp-agent-autoscaler

ENV HOME /home/azp-agent-autoscaler

COPY --from=build --chown=1000 /go/bin/azp-agent-autoscaler /bin/azp-agent-autoscaler

COPY --from=ca-certificates /usr/share/ca-certificates /usr/share/ca-certificates
COPY --from=ca-certificates /etc/ca-certificates.conf /etc/ca-certificates.conf
COPY --from=ca-certificates /etc/ssl/certs /etc/ssl/certs

ENTRYPOINT ["/bin/azp-agent-autoscaler"]
CMD []
