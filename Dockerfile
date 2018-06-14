FROM golang:1.9-alpine3.7 as builder

COPY . /go/src/github.com/chrisurwin/rancher-rebalancer

WORKDIR /go/src/github.com/chrisurwin/rancher-rebalancer/cmd/rebalance

RUN apk add --update --no-cache git gcc musl-dev && \
    go get ./... && \
    CGO_ENABLED=0 go build -a -ldflags '-extldflags "-static"' -o /opt/bin/rancher-rebalancer .

FROM centurylink/ca-certs

COPY --from=builder /opt/bin/rancher-rebalancer /opt/bin/rancher-rebalancer

WORKDIR /opt/bin

ENTRYPOINT ["./rancher-rebalancer"]
