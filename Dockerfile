FROM golang:1.12-alpine AS base

WORKDIR /src

COPY main.go /src/

COPY pkg /src/

RUN go build -o /app ./...

FROM scratch

COPY --from=base /app /bin/azp-agent-autoscaler

RUN chmod +x /bin/azp-agent-autoscaler

CMD ["/bin/azp-agent-autoscaler"]
