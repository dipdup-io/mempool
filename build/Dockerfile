# ---------------------------------------------------------------------
#  The first stage container, for building the application
# ---------------------------------------------------------------------
FROM golang:1.21-alpine as builder

ENV CGO_ENABLED=1
ENV GO111MODULE=on
ENV GOOS=linux

RUN apk --no-cache add ca-certificates
RUN apk add --update git musl-dev gcc build-base

RUN mkdir -p $GOPATH/src/github.com/dipdup-net/mempool/

COPY ./go.* $GOPATH/src/github.com/dipdup-net/mempool/
WORKDIR $GOPATH/src/github.com/dipdup-net/mempool
RUN go mod download

COPY cmd/mempool cmd/mempool

WORKDIR $GOPATH/src/github.com/dipdup-net/mempool/cmd/mempool/
RUN go build -ldflags "-linkmode external -extldflags '-static' -s -w" -a -installsuffix cgo -o /go/bin/dipdup-mempool .

# ---------------------------------------------------------------------
#  The second stage container, for running the application
# ---------------------------------------------------------------------
FROM scratch

WORKDIR /app/mempool

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /go/bin/dipdup-mempool /go/bin/dipdup-mempool
COPY ./build/dipdup.*.yml ./
COPY ./cmd/mempool/views/*.sql ./views/
COPY ./cmd/mempool/graphql ./graphql

ENTRYPOINT ["/go/bin/dipdup-mempool"]
