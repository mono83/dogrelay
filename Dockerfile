FROM golang:1.14-alpine as builder
WORKDIR /go/src/github.com/mono83/dogrelay
COPY . .
RUN apk add --no-cache build-base git \
    && go get ./... \
    && go test ./... \
    && CGO_ENABLED=0 go build -o release/dogrelay main.go


FROM alpine:3.12
COPY --from=builder /go/src/github.com/mono83/dogrelay/release/dogrelay /app/dogrelay
WORKDIR /app
