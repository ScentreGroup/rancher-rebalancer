FROM golang:1.12-alpine3.9 as builder

COPY . /go/src/github.com/scentregroup/rancher-rebalancer

WORKDIR /go/src/github.com/scentregroup/rancher-rebalancer/cmd/rebalance

RUN apk add --update --no-cache git gcc musl-dev && \
    go get ./... && \
    CGO_ENABLED=0 go build -a -ldflags '-extldflags "-static"' -o /opt/bin/rancher-rebalancer .

FROM centurylink/ca-certs

COPY --from=builder /opt/bin/rancher-rebalancer /opt/bin/rancher-rebalancer

WORKDIR /opt/bin

ENTRYPOINT ["./rancher-rebalancer"]
