FROM golang:1.11.5-alpine3.9 as builder
WORKDIR /go/src/github.com/mono83/dogrelay
COPY . .
RUN apk add --no-cache make git \
    && make release-docker


FROM alpine:3.9
COPY --from=builder /go/src/github.com/mono83/dogrelay/release/dogrelay-linux64 /app/dogrelay
WORKDIR /app
