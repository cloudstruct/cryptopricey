# syntax=docker/dockerfile:1

## Build
FROM golang:1.17-alpine as build

WORKDIR /build

COPY go.mod /build/go.mod
COPY go.sum /build/go.sum

RUN go mod download \
  && mkdir -p /build/assets

COPY *.go /build/
COPY assets/homeView.json /build/assets/homeView.json

RUN env CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /build/cryptopricey

## Deploy
# hadolint ignore=DL3006
FROM alpine:3

WORKDIR /

COPY --from=build /build/cryptopricey /cryptopricey

RUN addgroup -g 1000 -S cryptopricey && \
    adduser -u 1000 -S cryptopricey -G cryptopricey

RUN chmod 0755 /cryptopricey

USER cryptopricey

EXPOSE 8080

CMD ["./cryptopricey"]
