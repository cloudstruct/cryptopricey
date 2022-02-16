# syntax=docker/dockerfile:1

## Build
FROM golang:1.17-alpine as build

WORKDIR /app

COPY go.mod ./
COPY go.sum ./

RUN go mod download \
  && mkdir /app/assets

COPY *.go ./
COPY assets/* /app/assets/

RUN go build -o /cryptopricey

## Deploy
# hadolint ignore=DL3006
FROM debian:11-slim

WORKDIR /

RUN mkdir /app \
  && chmod 0755 /app

COPY --from=build /cryptopricey /app/cryptopricey

RUN chmod 0755 /app/cryptopricey

EXPOSE 8080

USER nonroot:nonroot

ENTRYPOINT ["/app/cryptopricey"]
