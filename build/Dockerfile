# ---------------------------------------------------------------------
#  The first stage container, for building the application
# ---------------------------------------------------------------------
FROM golang:1.15-alpine as builder

ENV CGO_ENABLED=0
ENV GO111MODULE=on
ENV GOOS=linux

RUN apk --no-cache add ca-certificates
RUN apk add --update git

RUN mkdir -p $GOPATH/src/github.com/dipdup-net/mempool/

COPY ./go.* $GOPATH/src/github.com/dipdup-net/mempool/
WORKDIR $GOPATH/src/github.com/dipdup-net/mempool
RUN go mod download

COPY cmd/mempool cmd/mempool
COPY internal internal

WORKDIR $GOPATH/src/github.com/dipdup-net/mempool/cmd/mempool/
RUN go build -a -installsuffix cgo -o /go/bin/dipdup-mempool .

# ---------------------------------------------------------------------
#  The second stage container, for running the application
# ---------------------------------------------------------------------
FROM scratch

WORKDIR /app/mempool

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /go/bin/dipdup-mempool /go/bin/dipdup-mempool
COPY config.yaml /app/mempool/

ENTRYPOINT ["/go/bin/dipdup-mempool"]
